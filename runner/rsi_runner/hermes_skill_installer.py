from __future__ import annotations

import argparse
from contextlib import contextmanager
from dataclasses import dataclass
from datetime import datetime, timezone
import fcntl
import hashlib
import json
import logging
import os
from pathlib import Path
import re
import shutil
import subprocess
import time
from typing import Any, Iterator

from rsi_runner.hermes_skill_exporter import (
    GitHubAuth,
    _atomic_json,
    _env,
    _is_relative_to,
    _normalize_private_key,
    _sha256_bytes,
    _sha256_file,
)


logger = logging.getLogger(__name__)

INSTALLER_STATE_VERSION = 1
MANIFEST_NAME = "manifest.json"
EMPTY_TREE_HASH = _sha256_bytes(b"[]")
SOURCE_PATH_PREFIX = "hermes/skills/"


class InstallerConfigError(ValueError):
    pass


class InstallerError(RuntimeError):
    pass


SECRET_PATTERNS = [
    re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH |)PRIVATE KEY-----"),
    re.compile(r"\bsk-or-v1-[A-Za-z0-9_-]{24,}\b"),
    re.compile(r"\bsk-[A-Za-z0-9_-]{24,}\b"),
    re.compile(r"\bghp_[A-Za-z0-9_]{24,}\b"),
    re.compile(r"\bgithub_pat_[A-Za-z0-9_]{24,}\b"),
    re.compile(r"\bxox[baprs]-[A-Za-z0-9-]{24,}\b"),
    re.compile(
        r"(?i)\b(?:api[_-]?key|token|secret|password|private[_-]?key)\s*[:=]\s*"
        r"[\"']?(?!<|\\$|\$\{|REDACTED\b|EXAMPLE\b|YOUR_|VAULT_|vault:)"
        r"[A-Za-z0-9_./+=:-]{20,}"
    ),
]


def _utc_now() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z")


def _positive_float_env(name: str, default: float) -> float:
    raw = _env(name, str(default))
    try:
        value = float(raw)
    except ValueError as exc:
        raise InstallerConfigError(f"{name} must be a positive number") from exc
    if value <= 0:
        raise InstallerConfigError(f"{name} must be a positive number")
    return value


def _bool_env(name: str, default: bool = False) -> bool:
    raw = _env(name)
    if not raw:
        return default
    value = raw.lower()
    if value in {"1", "true", "t", "yes", "y", "on"}:
        return True
    if value in {"0", "false", "f", "no", "n", "off"}:
        return False
    raise InstallerConfigError(f"{name} must be a boolean")


def _sanitize_relative_path(value: str, *, label: str) -> str:
    raw = value.strip().strip("/")
    if not raw:
        raise InstallerConfigError(f"{label} must not be empty")
    path = Path(raw)
    if path.is_absolute() or ".." in path.parts:
        raise InstallerConfigError(f"{label} must be a safe relative path")
    return path.as_posix()


def _default_target_relative_path(source_path: str) -> str:
    normalized = source_path.strip("/")
    if normalized.startswith(SOURCE_PATH_PREFIX):
        return _sanitize_relative_path(normalized[len(SOURCE_PATH_PREFIX) :], label="target relative path")
    return _sanitize_relative_path(Path(normalized).name, label="target relative path")


def _load_json(path: Path) -> dict[str, Any]:
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError as exc:
        raise InstallerError(f"missing skill manifest: {path}") from exc
    except json.JSONDecodeError as exc:
        raise InstallerError(f"invalid JSON skill manifest: {path}: {exc}") from exc
    if not isinstance(payload, dict):
        raise InstallerError(f"skill manifest must be a JSON object: {path}")
    return payload


@dataclass(frozen=True)
class TreeFile:
    relative_path: str
    absolute_path: Path
    sha256: str
    size: int


@dataclass(frozen=True)
class TreeSnapshot:
    tree_hash: str
    files: list[TreeFile]

    def manifest(self) -> list[dict[str, Any]]:
        return [{"path": item.relative_path, "sha256": item.sha256, "size": item.size} for item in self.files]


def build_tree_snapshot(root: Path, *, excluded_relative_paths: set[str] | None = None) -> TreeSnapshot:
    excluded = excluded_relative_paths or set()
    if not root.exists():
        files: list[TreeFile] = []
    elif not root.is_dir():
        raise InstallerError(f"tree root is not a directory: {root}")
    else:
        resolved_root = root.resolve()
        files = []
        for path in sorted(resolved_root.rglob("*")):
            if path.is_symlink():
                raise InstallerError(f"skill pack must not contain symlinks: {path}")
            if not path.is_file():
                continue
            rel = path.relative_to(resolved_root).as_posix()
            if rel in excluded or rel == ".bundled_manifest" or rel.startswith(".hub/"):
                continue
            stat = path.stat()
            files.append(TreeFile(relative_path=rel, absolute_path=path, sha256=_sha256_file(path), size=stat.st_size))
    manifest_bytes = json.dumps(
        [{"path": item.relative_path, "sha256": item.sha256, "size": item.size} for item in files],
        ensure_ascii=True,
        sort_keys=True,
        separators=(",", ":"),
    ).encode("utf-8")
    return TreeSnapshot(tree_hash=_sha256_bytes(manifest_bytes), files=files)


def scan_for_secret_values(root: Path, *, excluded_relative_paths: set[str] | None = None) -> None:
    excluded = excluded_relative_paths or set()
    if not root.exists():
        raise InstallerError(f"source skill root does not exist: {root}")
    for path in sorted(root.rglob("*")):
        if path.is_symlink():
            raise InstallerError(f"skill pack must not contain symlinks: {path}")
        if not path.is_file():
            continue
        rel = path.relative_to(root).as_posix()
        if rel in excluded:
            continue
        try:
            text = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        for pattern in SECRET_PATTERNS:
            if pattern.search(text):
                raise InstallerError(f"secret-like value detected in skill pack: {rel}")


@dataclass(frozen=True)
class SkillPackManifest:
    schema_version: int
    pack_version: str
    source_commit: str
    expected_previous_live_tree_hash: str
    target_skill_tree_hash: str

    @classmethod
    def load(cls, source_root: Path) -> "SkillPackManifest":
        payload = _load_json(source_root / MANIFEST_NAME)
        try:
            schema_version = int(payload.get("schema_version"))
        except (TypeError, ValueError) as exc:
            raise InstallerError("manifest schema_version must be an integer") from exc
        if schema_version != 1:
            raise InstallerError(f"unsupported manifest schema_version: {schema_version}")
        pack_version = str(payload.get("pack_version") or "").strip()
        target_hash = str(payload.get("target_skill_tree_hash") or "").strip()
        if not pack_version:
            raise InstallerError("manifest pack_version is required")
        if not re.fullmatch(r"[0-9a-f]{64}", target_hash):
            raise InstallerError("manifest target_skill_tree_hash must be a sha256 hex digest")
        expected = str(payload.get("expected_previous_live_tree_hash") or "").strip()
        if expected and not re.fullmatch(r"[0-9a-f]{64}", expected):
            raise InstallerError("manifest expected_previous_live_tree_hash must be empty or a sha256 hex digest")
        return cls(
            schema_version=schema_version,
            pack_version=pack_version,
            source_commit=str(payload.get("source_commit") or "").strip(),
            expected_previous_live_tree_hash=expected,
            target_skill_tree_hash=target_hash,
        )


@dataclass(frozen=True)
class GitHubAuthConfig:
    github_token: str = ""
    github_app_id: str = ""
    github_app_installation_id: str = ""
    github_app_private_key: str = ""


class InstallerGitHubAuth(GitHubAuth):
    def __init__(self, config: GitHubAuthConfig) -> None:
        self.config = config  # type: ignore[assignment]

    def token(self) -> str:
        if self.config.github_token:
            return self.config.github_token
        return super().token()


@dataclass(frozen=True)
class InstallerConfig:
    source_repo: str
    source_ref: str
    source_path: str
    source_local_path: Path | None
    skills_root: Path
    target_relative_path: str
    state_root: Path
    lock_path: Path
    dry_run: bool
    validate_only: bool
    lock_timeout_seconds: float
    github_auth: GitHubAuthConfig

    def __post_init__(self) -> None:
        skills_root = self.skills_root.resolve()
        target = (skills_root / self.target_relative_path).resolve()
        if not _is_relative_to(target, skills_root) or target == skills_root:
            raise InstallerConfigError("target skill path must be a non-root path under RSI_HERMES_SKILL_INSTALLER_SKILLS_ROOT")
        if self.state_root.resolve() == skills_root or _is_relative_to(self.state_root.resolve(), skills_root):
            raise InstallerConfigError("RSI_HERMES_SKILL_INSTALLER_STATE_ROOT must not be inside the skills root")
        if not _is_relative_to(self.lock_path.resolve(), self.skills_root.parent.resolve()):
            raise InstallerConfigError("RSI_HERMES_SKILL_INSTALLER_LOCK_PATH must stay under HERMES_HOME")
        if not self.source_local_path and (not self.source_repo or not self.source_ref):
            raise InstallerConfigError("source repo and source ref are required unless source local path is set")

    @property
    def target_root(self) -> Path:
        return self.skills_root / self.target_relative_path

    @property
    def state_path(self) -> Path:
        return self.state_root / "state.json"

    @classmethod
    def from_env(cls, args: argparse.Namespace) -> "InstallerConfig":
        hermes_home = Path(_env("HERMES_HOME", "/var/lib/hermes")).expanduser()
        source_path = _env("RSI_HERMES_SKILL_INSTALLER_SOURCE_PATH", "hermes/skills/story-company")
        target_relative_path = _env("RSI_HERMES_SKILL_INSTALLER_TARGET_RELATIVE_PATH") or _default_target_relative_path(source_path)
        source_local_raw = args.source_local_path or _env("RSI_HERMES_SKILL_INSTALLER_SOURCE_LOCAL_PATH")
        return cls(
            source_repo=_env("RSI_HERMES_SKILL_INSTALLER_SOURCE_REPO", "piplabs/rsi-agent-platform"),
            source_ref=args.source_ref or _env("RSI_HERMES_SKILL_INSTALLER_SOURCE_REF"),
            source_path=source_path,
            source_local_path=Path(source_local_raw).expanduser() if source_local_raw else None,
            skills_root=Path(_env("RSI_HERMES_SKILL_INSTALLER_SKILLS_ROOT", str(hermes_home / "skills"))).expanduser(),
            target_relative_path=target_relative_path,
            state_root=Path(_env("RSI_HERMES_SKILL_INSTALLER_STATE_ROOT", str(hermes_home / "skill-installer"))).expanduser(),
            lock_path=Path(_env("RSI_HERMES_SKILL_INSTALLER_LOCK_PATH", str(hermes_home / ".locks/skills.lock"))).expanduser(),
            dry_run=args.dry_run or _bool_env("RSI_HERMES_SKILL_INSTALLER_DRY_RUN", False),
            validate_only=args.validate_only,
            lock_timeout_seconds=_positive_float_env("RSI_HERMES_SKILL_INSTALLER_LOCK_TIMEOUT_SECONDS", 60.0),
            github_auth=GitHubAuthConfig(
                github_token=(
                    _env("RSI_HERMES_SKILL_INSTALLER_GITHUB_TOKEN")
                    or _env("RSI_GITHUB_TOKEN")
                    or _env("GITHUB_TOKEN")
                ),
                github_app_id=_env("RSI_GITHUB_APP_ID"),
                github_app_installation_id=_env("RSI_GITHUB_APP_INSTALLATION_ID"),
                github_app_private_key=_normalize_private_key(_env("RSI_GITHUB_APP_PRIVATE_KEY")),
            ),
        )


@contextmanager
def exclusive_lock(path: Path, timeout_seconds: float) -> Iterator[None]:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a+", encoding="utf-8") as handle:
        deadline = time.monotonic() + timeout_seconds
        while True:
            try:
                fcntl.flock(handle.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
                break
            except BlockingIOError:
                if time.monotonic() >= deadline:
                    raise InstallerError(f"timed out waiting for skill lock: {path}")
                time.sleep(0.25)
        try:
            yield
        finally:
            fcntl.flock(handle.fileno(), fcntl.LOCK_UN)


class SkillSourceCheckout:
    def __init__(self, config: InstallerConfig) -> None:
        self.config = config
        self.checkout: Path | None = None

    def __enter__(self) -> Path:
        if self.config.source_local_path:
            source_root = (self.config.source_local_path / self.config.source_path).resolve()
            if not source_root.is_dir():
                raise InstallerError(f"source skill path does not exist: {source_root}")
            return source_root
        token = InstallerGitHubAuth(self.config.github_auth).token()
        checkout = self.config.state_root / "checkouts" / self.config.source_ref[:12]
        if checkout.exists():
            shutil.rmtree(checkout)
        checkout.parent.mkdir(parents=True, exist_ok=True)
        checkout.mkdir(parents=True)
        remote = f"https://x-access-token:{token}@github.com/{self.config.source_repo}.git"
        self._git(checkout, "init")
        self._git(checkout, "remote", "add", "origin", remote)
        self._git(checkout, "fetch", "--depth=1", "origin", self.config.source_ref)
        self._git(checkout, "checkout", "--detach", "FETCH_HEAD")
        self.checkout = checkout
        source_root = (checkout / self.config.source_path).resolve()
        if not source_root.is_dir():
            raise InstallerError(f"source skill path does not exist at {self.config.source_ref}: {self.config.source_path}")
        return source_root

    def __exit__(self, _exc_type: object, _exc: object, _tb: object) -> None:
        if self.checkout is not None:
            shutil.rmtree(self.checkout, ignore_errors=True)

    @staticmethod
    def _git(cwd: Path, *args: str) -> None:
        env = dict(os.environ)
        env["GIT_TERMINAL_PROMPT"] = "0"
        completed = subprocess.run(["git", *args], cwd=str(cwd), env=env, text=True, capture_output=True, check=False)
        if completed.returncode != 0:
            redact_pattern = re.compile(r"x-access-token:[^@\s]+")
            redacted = redact_pattern.sub("x-access-token:***", completed.stderr or completed.stdout)
            sanitized_args = [redact_pattern.sub("x-access-token:***", arg) for arg in args]
            raise InstallerError(f"git {' '.join(sanitized_args)} failed: {redacted}")


class SkillInstaller:
    def __init__(self, config: InstallerConfig) -> None:
        self.config = config

    def run(self) -> dict[str, Any]:
        with SkillSourceCheckout(self.config) as source_root:
            manifest = SkillPackManifest.load(source_root)
            source_snapshot = self._validate_source(source_root, manifest)
            if self.config.validate_only:
                return self._result("validated", manifest, source_snapshot, live_snapshot=None, changed=False)
            with exclusive_lock(self.config.lock_path, self.config.lock_timeout_seconds):
                live_snapshot = build_tree_snapshot(self.config.target_root)
                if live_snapshot.tree_hash == source_snapshot.tree_hash:
                    self._record_state(manifest, source_snapshot, live_snapshot, status="unchanged")
                    return self._result("unchanged", manifest, source_snapshot, live_snapshot=live_snapshot, changed=False)
                expected_hash = self._expected_live_hash(manifest)
                if live_snapshot.tree_hash != expected_hash:
                    return self._fail_diverged(manifest, source_snapshot, live_snapshot, expected_hash)
                if not self.config.dry_run:
                    self._install_tree(source_root)
                    installed_snapshot = build_tree_snapshot(self.config.target_root)
                    if installed_snapshot.tree_hash != source_snapshot.tree_hash:
                        raise InstallerError(
                            f"installed tree hash {installed_snapshot.tree_hash} did not match source {source_snapshot.tree_hash}"
                        )
                    self._record_state(manifest, source_snapshot, live_snapshot, status="installed")
                    live_snapshot = installed_snapshot
                return self._result("dry_run" if self.config.dry_run else "installed", manifest, source_snapshot, live_snapshot=live_snapshot, changed=True)

    def _validate_source(self, source_root: Path, manifest: SkillPackManifest) -> TreeSnapshot:
        scan_for_secret_values(source_root, excluded_relative_paths={MANIFEST_NAME})
        source_snapshot = build_tree_snapshot(source_root, excluded_relative_paths={MANIFEST_NAME})
        if source_snapshot.tree_hash != manifest.target_skill_tree_hash:
            raise InstallerError(
                "manifest target_skill_tree_hash does not match source pack: "
                f"manifest={manifest.target_skill_tree_hash} computed={source_snapshot.tree_hash}"
            )
        if manifest.source_commit and manifest.source_commit not in {"auto", self.config.source_ref}:
            raise InstallerError(
                f"manifest source_commit {manifest.source_commit} does not match requested source ref {self.config.source_ref}"
            )
        return source_snapshot

    def _expected_live_hash(self, manifest: SkillPackManifest) -> str:
        if manifest.expected_previous_live_tree_hash:
            return manifest.expected_previous_live_tree_hash
        state = self._load_state()
        if state.get("target_relative_path") == self.config.target_relative_path and state.get("last_installed_target_hash"):
            return str(state["last_installed_target_hash"])
        return EMPTY_TREE_HASH

    def _load_state(self) -> dict[str, Any]:
        try:
            payload = json.loads(self.config.state_path.read_text(encoding="utf-8"))
        except (FileNotFoundError, json.JSONDecodeError):
            return {}
        return payload if isinstance(payload, dict) else {}

    def _record_state(
        self,
        manifest: SkillPackManifest,
        source_snapshot: TreeSnapshot,
        previous_live_snapshot: TreeSnapshot,
        *,
        status: str,
    ) -> None:
        state = {
            "state_version": INSTALLER_STATE_VERSION,
            "last_status": status,
            "last_installed_at": _utc_now(),
            "source_repo": self.config.source_repo,
            "source_ref": self.config.source_ref,
            "source_path": self.config.source_path,
            "target_relative_path": self.config.target_relative_path,
            "pack_version": manifest.pack_version,
            "previous_live_tree_hash": previous_live_snapshot.tree_hash,
            "last_installed_target_hash": source_snapshot.tree_hash,
            "file_count": len(source_snapshot.files),
            "files": source_snapshot.manifest(),
        }
        _atomic_json(self.config.state_path, state)
        ledger = self.config.state_root / "ledger" / f"{datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')}-{source_snapshot.tree_hash[:12]}.json"
        _atomic_json(ledger, state)

    def _install_tree(self, source_root: Path) -> None:
        target_root = self.config.target_root.resolve()
        target_root.parent.mkdir(parents=True, exist_ok=True)
        operation_id = f"{int(time.time())}-{os.getpid()}-{hashlib.sha256(str(target_root).encode()).hexdigest()[:8]}"
        staging_parent = self.config.state_root / "staging" / operation_id
        backup_parent = self.config.state_root / "backups" / operation_id
        staging_target = staging_parent / "target"
        backup_target = backup_parent / "target"
        try:
            staging_target.mkdir(parents=True)
            self._copy_source_tree(source_root, staging_target)
            if target_root.exists():
                backup_target.parent.mkdir(parents=True, exist_ok=True)
                os.replace(target_root, backup_target)
            os.replace(staging_target, target_root)
            shutil.rmtree(backup_parent, ignore_errors=True)
        except Exception:
            if not target_root.exists() and backup_target.exists():
                os.replace(backup_target, target_root)
            raise
        finally:
            shutil.rmtree(staging_parent, ignore_errors=True)

    @staticmethod
    def _copy_source_tree(source_root: Path, target_root: Path) -> None:
        for item in sorted(source_root.rglob("*")):
            rel = item.relative_to(source_root)
            if rel.as_posix() == MANIFEST_NAME:
                continue
            if item.is_symlink():
                raise InstallerError(f"skill pack must not contain symlinks: {item}")
            target = target_root / rel
            if item.is_dir():
                target.mkdir(parents=True, exist_ok=True)
            elif item.is_file():
                target.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, target)

    def _fail_diverged(
        self,
        manifest: SkillPackManifest,
        source_snapshot: TreeSnapshot,
        live_snapshot: TreeSnapshot,
        expected_hash: str,
    ) -> dict[str, Any]:
        report = self._result("diverged", manifest, source_snapshot, live_snapshot=live_snapshot, changed=False)
        report["ok"] = False
        report["expected_live_tree_hash"] = expected_hash
        report_path = self.config.state_root / "conflicts" / f"{datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')}-{live_snapshot.tree_hash[:12]}.json"
        _atomic_json(report_path, report)
        raise InstallerError(
            "live skill tree diverged; refusing install: "
            f"target={self.config.target_relative_path} expected={expected_hash} actual={live_snapshot.tree_hash} "
            f"conflict_report={report_path}"
        )

    def _result(
        self,
        status: str,
        manifest: SkillPackManifest,
        source_snapshot: TreeSnapshot,
        *,
        live_snapshot: TreeSnapshot | None,
        changed: bool,
    ) -> dict[str, Any]:
        return {
            "ok": True,
            "status": status,
            "changed": changed,
            "completed_at": _utc_now(),
            "dry_run": self.config.dry_run,
            "source_repo": self.config.source_repo,
            "source_ref": self.config.source_ref,
            "source_path": self.config.source_path,
            "target_relative_path": self.config.target_relative_path,
            "pack_version": manifest.pack_version,
            "target_skill_tree_hash": source_snapshot.tree_hash,
            "live_tree_hash": live_snapshot.tree_hash if live_snapshot else "",
            "file_count": len(source_snapshot.files),
        }


def configure_logging() -> None:
    level_name = _env("RSI_HERMES_SKILL_INSTALLER_LOG_LEVEL", _env("RSI_RUNNER_LOG_LEVEL", "INFO")).upper()
    logging.basicConfig(level=getattr(logging, level_name, logging.INFO), format="%(asctime)s %(levelname)s %(name)s %(message)s")


def main() -> None:
    parser = argparse.ArgumentParser(description="Install reviewed Story Hermes skills into a live Hermes home")
    parser.add_argument("--source-ref", default="", help="Git commit/ref to install")
    parser.add_argument("--source-local-path", default="", help="Local checkout root for validation/tests")
    parser.add_argument("--dry-run", action="store_true", help="Validate and compare hashes without writing skills")
    parser.add_argument("--validate-only", action="store_true", help="Validate the source pack and manifest without touching live skills")
    args = parser.parse_args()
    configure_logging()
    config = InstallerConfig.from_env(args)
    try:
        result = SkillInstaller(config).run()
    except Exception as exc:
        logger.exception("Hermes skill install failed")
        print(json.dumps({"ok": False, "error": str(exc)}, ensure_ascii=True, sort_keys=True))
        raise SystemExit(1) from exc
    print(json.dumps(result, ensure_ascii=True, sort_keys=True))


if __name__ == "__main__":
    main()
