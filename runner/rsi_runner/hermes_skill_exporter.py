from __future__ import annotations

import argparse
import base64
from dataclasses import dataclass, field
from datetime import datetime, timezone
import hashlib
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import logging
import os
from pathlib import Path
import shutil
import signal
import sqlite3
import subprocess
import tempfile
import threading
import time
from typing import Any
from urllib import error, request


logger = logging.getLogger(__name__)

EXPORT_ROOT_PREFIX = "hermes/exported-skills"
STATE_VERSION = 1
MAX_EXPORTED_HASHES = 500


class ExporterConfigError(ValueError):
    pass


class ExporterError(RuntimeError):
    pass


def _env(name: str, default: str = "") -> str:
    return os.getenv(name, default).strip()


def _bool_env(name: str, default: bool = False) -> bool:
    raw = _env(name)
    if not raw:
        return default
    value = raw.lower()
    if value in {"1", "true", "t", "yes", "y", "on"}:
        return True
    if value in {"0", "false", "f", "no", "n", "off"}:
        return False
    raise ExporterConfigError(f"{name} must be a boolean")


def _positive_int_env(name: str, default: int) -> int:
    raw = _env(name, str(default))
    try:
        value = int(raw)
    except ValueError as exc:
        raise ExporterConfigError(f"{name} must be a positive integer") from exc
    if value <= 0:
        raise ExporterConfigError(f"{name} must be a positive integer")
    return value


def _utc_now() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z")


def _safe_slug(value: str) -> str:
    out = []
    for char in value.lower():
        if char.isalnum() or char in {"-", "_", "/"}:
            out.append(char)
        elif char in {".", " "}:
            out.append("-")
    return "".join(out).strip("-/") or "stage"


def _is_relative_to(path: Path, parent: Path) -> bool:
    try:
        path.resolve().relative_to(parent.resolve())
        return True
    except ValueError:
        return False


def _load_json(path: Path, default: dict[str, Any]) -> dict[str, Any]:
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (FileNotFoundError, json.JSONDecodeError):
        return dict(default)
    return payload if isinstance(payload, dict) else dict(default)


def _atomic_json(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temp_name = tempfile.mkstemp(prefix=path.name + ".", suffix=".tmp", dir=str(path.parent))
    temp_path = Path(temp_name)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, ensure_ascii=True, indent=2, sort_keys=True)
            handle.write("\n")
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(temp_path, path)
    except Exception:
        temp_path.unlink(missing_ok=True)
        raise


def _sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _normalize_private_key(raw: str) -> str:
    value = raw.strip()
    if "\\n" in value and "\n" not in value:
        value = value.replace("\\n", "\n")
    if "BEGIN" in value:
        return value
    try:
        decoded = base64.b64decode(value).decode("utf-8")
    except Exception:
        return value
    return decoded if "BEGIN" in decoded else value


@dataclass(frozen=True)
class ExporterConfig:
    host: str
    port: int
    enabled: bool
    hermes_home: Path
    skills_root: Path
    state_root: Path
    sync_interval_seconds: int
    git_owner: str
    git_repo: str
    git_base_branch: str
    export_env: str
    branch_prefix: str
    pr_mode: str
    pod_name: str
    github_token: str = ""
    github_app_id: str = ""
    github_app_installation_id: str = ""
    github_app_private_key: str = ""

    def __post_init__(self) -> None:
        hermes_home = self.hermes_home.resolve()
        skills_root = self.skills_root.resolve()
        state_root = self.state_root.resolve()
        if state_root == skills_root or _is_relative_to(state_root, skills_root):
            raise ExporterConfigError("RSI_HERMES_SKILL_EXPORTER_STATE_ROOT must not be inside skills root")
        if skills_root == state_root or _is_relative_to(skills_root, state_root):
            raise ExporterConfigError("RSI_HERMES_SKILL_EXPORTER_SKILLS_ROOT must not be inside state root")
        if state_root == hermes_home:
            raise ExporterConfigError("RSI_HERMES_SKILL_EXPORTER_STATE_ROOT must not equal HERMES_HOME")
        if self.pr_mode != "per_change":
            raise ExporterConfigError("RSI_HERMES_SKILL_EXPORTER_PR_MODE currently supports only per_change")
        if not self.git_owner or not self.git_repo:
            raise ExporterConfigError("Git owner and repo are required")

    @property
    def git_repository(self) -> str:
        return f"{self.git_owner}/{self.git_repo}"

    @property
    def export_root(self) -> str:
        return f"{EXPORT_ROOT_PREFIX}/{_safe_slug(self.export_env)}"

    @classmethod
    def from_env(cls) -> "ExporterConfig":
        legacy = sorted(name for name in os.environ if name.startswith("RSI_HERMES_LEARNER_"))
        if legacy:
            raise ExporterConfigError("legacy Hermes learner env is not supported: " + ", ".join(legacy))
        hermes_home = Path(_env("HERMES_HOME", "/var/lib/hermes")).expanduser()
        github_token = (
            _env("RSI_HERMES_SKILL_EXPORTER_GITHUB_TOKEN")
            or _env("RSI_GITHUB_TOKEN")
            or _env("GITHUB_TOKEN")
        )
        return cls(
            host=_env("RSI_HERMES_SKILL_EXPORTER_HOST", _env("RSI_RUNNER_HOST", "0.0.0.0")),
            port=_positive_int_env("RSI_HERMES_SKILL_EXPORTER_PORT", 8091),
            enabled=_bool_env("RSI_HERMES_SKILL_EXPORTER_ENABLED", True),
            hermes_home=hermes_home,
            skills_root=Path(_env("RSI_HERMES_SKILL_EXPORTER_SKILLS_ROOT", str(hermes_home / "skills"))).expanduser(),
            state_root=Path(_env("RSI_HERMES_SKILL_EXPORTER_STATE_ROOT", str(hermes_home / "skill-exporter"))).expanduser(),
            sync_interval_seconds=_positive_int_env("RSI_HERMES_SKILL_EXPORTER_SYNC_INTERVAL_SECONDS", 300),
            git_owner=_env("RSI_HERMES_SKILL_EXPORTER_GIT_OWNER", _env("RSI_GITHUB_OWNER", "piplabs")),
            git_repo=_env("RSI_HERMES_SKILL_EXPORTER_GIT_REPO", "rsi-agent-platform"),
            git_base_branch=_env("RSI_HERMES_SKILL_EXPORTER_GIT_BASE_BRANCH", "main"),
            export_env=_env("RSI_HERMES_SKILL_EXPORTER_EXPORT_ENV", "stage"),
            branch_prefix=_env("RSI_HERMES_SKILL_EXPORTER_BRANCH_PREFIX", "hermes/skill-export"),
            pr_mode=_env("RSI_HERMES_SKILL_EXPORTER_PR_MODE", "per_change"),
            pod_name=_env("POD_NAME", _env("HOSTNAME", "unknown")),
            github_token=github_token,
            github_app_id=_env("RSI_GITHUB_APP_ID"),
            github_app_installation_id=_env("RSI_GITHUB_APP_INSTALLATION_ID"),
            github_app_private_key=_normalize_private_key(_env("RSI_GITHUB_APP_PRIVATE_KEY")),
        )


@dataclass(frozen=True)
class SkillFile:
    relative_path: str
    absolute_path: Path
    sha256: str
    size: int


@dataclass(frozen=True)
class SkillSnapshot:
    tree_hash: str
    files: list[SkillFile]
    recorded_at: str

    def file_manifest(self) -> list[dict[str, Any]]:
        return [
            {"path": item.relative_path, "sha256": item.sha256, "size": item.size}
            for item in self.files
        ]


def build_skill_snapshot(skills_root: Path) -> SkillSnapshot:
    root = skills_root.resolve()
    if not root.exists():
        raise ExporterError(f"skills root does not exist: {root}")
    if not root.is_dir():
        raise ExporterError(f"skills root is not a directory: {root}")
    files: list[SkillFile] = []
    for path in sorted(root.rglob("*")):
        if not path.is_file():
            continue
        rel = path.relative_to(root).as_posix()
        if rel == ".bundled_manifest" or rel.startswith(".hub/"):
            continue
        stat = path.stat()
        files.append(
            SkillFile(
                relative_path=rel,
                absolute_path=path,
                sha256=_sha256_file(path),
                size=stat.st_size,
            )
        )
    manifest_bytes = json.dumps(
        [{"path": item.relative_path, "sha256": item.sha256, "size": item.size} for item in files],
        ensure_ascii=True,
        sort_keys=True,
        separators=(",", ":"),
    ).encode("utf-8")
    return SkillSnapshot(tree_hash=_sha256_bytes(manifest_bytes), files=files, recorded_at=_utc_now())


def validate_export_paths(export_root: str, paths: list[str]) -> None:
    prefix = export_root.rstrip("/") + "/"
    invalid = [path for path in paths if path != export_root and not path.startswith(prefix)]
    if invalid:
        raise ExporterError("export touches non-export paths: " + ", ".join(sorted(invalid)))


class SkillProvenanceReader:
    def __init__(self, hermes_home: Path) -> None:
        self.db_path = hermes_home / "state.db"

    def skill_manage_sessions(self, limit: int = 50) -> list[dict[str, Any]]:
        if not self.db_path.exists():
            return []
        try:
            with sqlite3.connect(f"file:{self.db_path}?mode=ro", uri=True) as connection:
                connection.row_factory = sqlite3.Row
                columns = {
                    row["name"]
                    for row in connection.execute("PRAGMA table_info(messages)").fetchall()
                    if "name" in row.keys()
                }
                if not {"session_id", "tool_name", "tool_calls", "content"}.issubset(columns):
                    return []
                rows = connection.execute(
                    """
                    SELECT session_id, timestamp, role, tool_name, content, tool_calls
                    FROM messages
                    WHERE lower(coalesce(tool_name, '')) LIKE '%skill%'
                       OR lower(coalesce(tool_calls, '')) LIKE '%skill_manage%'
                       OR lower(coalesce(content, '')) LIKE '%skill_manage%'
                    ORDER BY timestamp DESC
                    LIMIT ?
                    """,
                    (limit,),
                ).fetchall()
        except sqlite3.Error as exc:
            logger.warning("failed to read Hermes skill provenance: %s", exc)
            return []
        sessions: list[dict[str, Any]] = []
        seen: set[str] = set()
        for row in rows:
            session_id = str(row["session_id"] or "")
            if not session_id or session_id in seen:
                continue
            seen.add(session_id)
            sessions.append(
                {
                    "session_id": session_id,
                    "timestamp": row["timestamp"],
                    "tool_name": row["tool_name"],
                    "role": row["role"],
                }
            )
        return sessions


class GitHubAuth:
    def __init__(self, config: ExporterConfig) -> None:
        self.config = config

    def token(self) -> str:
        if self.config.github_token:
            return self.config.github_token
        if not (
            self.config.github_app_id
            and self.config.github_app_installation_id
            and self.config.github_app_private_key
        ):
            raise ExporterError("GitHub token or GitHub App credentials are required for skill export")
        return self._installation_token()

    def _installation_token(self) -> str:
        try:
            import jwt  # type: ignore
        except ImportError as exc:
            raise ExporterError("PyJWT[crypto] is required for GitHub App authentication") from exc
        now = int(time.time())
        app_jwt = jwt.encode(
            {"iat": now - 60, "exp": now + 540, "iss": self.config.github_app_id},
            self.config.github_app_private_key,
            algorithm="RS256",
        )
        url = (
            "https://api.github.com/app/installations/"
            f"{self.config.github_app_installation_id}/access_tokens"
        )
        payload = self._json_request(
            "POST",
            url,
            token=app_jwt,
            accept="application/vnd.github+json",
            body={},
        )
        token = str(payload.get("token") or "")
        if not token:
            raise ExporterError("GitHub App installation token response did not include token")
        return token

    @staticmethod
    def _json_request(
        method: str,
        url: str,
        *,
        token: str,
        accept: str = "application/vnd.github+json",
        body: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        data = None if body is None else json.dumps(body).encode("utf-8")
        req = request.Request(url, data=data, method=method)
        req.add_header("Accept", accept)
        req.add_header("Authorization", f"Bearer {token}")
        req.add_header("X-GitHub-Api-Version", "2022-11-28")
        if data is not None:
            req.add_header("Content-Type", "application/json")
        try:
            with request.urlopen(req, timeout=30) as response:
                raw = response.read().decode("utf-8")
        except error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            raise ExporterError(f"GitHub API {method} {url} failed: {exc.code} {detail}") from exc
        return json.loads(raw) if raw else {}


class GitSkillExporter:
    def __init__(self, config: ExporterConfig) -> None:
        self.config = config
        self.auth = GitHubAuth(config)

    def export(self, snapshot: SkillSnapshot, metadata: dict[str, Any]) -> dict[str, Any]:
        token = self.auth.token()
        branch = self._branch_name(snapshot)
        checkout = self.config.state_root / "checkouts" / snapshot.tree_hash[:12]
        if checkout.exists():
            shutil.rmtree(checkout)
        checkout.parent.mkdir(parents=True, exist_ok=True)
        checkout.mkdir()
        remote = f"https://x-access-token:{token}@github.com/{self.config.git_repository}.git"
        try:
            self._git(checkout, "init")
            self._git(checkout, "remote", "add", "origin", remote)
            self._git(checkout, "fetch", "--depth=1", "origin", self.config.git_base_branch)
            self._git(checkout, "checkout", "-b", branch, "FETCH_HEAD")
            self._write_export_tree(checkout, snapshot, metadata)
            self._git(checkout, "add", self.config.export_root)
            diff = self._git(checkout, "diff", "--cached", "--quiet", check=False)
            if diff.returncode == 0:
                return {"exported": False, "branch": branch, "reason": "no_git_diff"}
            self._git(checkout, "config", "user.name", "rsi-agent-platform-bot")
            self._git(checkout, "config", "user.email", "rsi-agent-platform-bot@users.noreply.github.com")
            message = f"Export Hermes skills {self.config.export_env} {snapshot.tree_hash[:12]}"
            self._git(checkout, "commit", "-m", message)
            self._git(checkout, "push", "--set-upstream", "origin", branch)
            pr = self._open_pr(token, branch, snapshot, metadata)
            return {"exported": True, "branch": branch, "pr_url": pr.get("html_url", ""), "pr_number": pr.get("number")}
        finally:
            shutil.rmtree(checkout, ignore_errors=True)

    def _branch_name(self, snapshot: SkillSnapshot) -> str:
        stamp = datetime.now(timezone.utc).strftime("%Y%m%d%H%M%S")
        return f"{self.config.branch_prefix}/{_safe_slug(self.config.export_env)}/{stamp}-{snapshot.tree_hash[:12]}"

    def _write_export_tree(self, checkout: Path, snapshot: SkillSnapshot, metadata: dict[str, Any]) -> None:
        export_root = checkout / self.config.export_root
        if export_root.exists():
            shutil.rmtree(export_root)
        for item in snapshot.files:
            target = export_root / "skills" / item.relative_path
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(item.absolute_path, target)
        metadata_path = export_root / "metadata.json"
        metadata_path.parent.mkdir(parents=True, exist_ok=True)
        metadata_path.write_text(json.dumps(metadata, ensure_ascii=True, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        exported_paths = [
            str(path.relative_to(checkout)).replace(os.sep, "/")
            for path in export_root.rglob("*")
            if path.is_file()
        ]
        validate_export_paths(self.config.export_root, exported_paths)

    def _open_pr(self, token: str, branch: str, snapshot: SkillSnapshot, metadata: dict[str, Any]) -> dict[str, Any]:
        title = f"Export Hermes skills from {self.config.export_env} ({snapshot.tree_hash[:12]})"
        body = "\n".join(
            [
                "Automated export of Hermes executor skill files for visibility.",
                "",
                f"- Environment: `{self.config.export_env}`",
                f"- Tree hash: `{snapshot.tree_hash}`",
                f"- File count: `{len(snapshot.files)}`",
                f"- Pod: `{metadata.get('pod_name', 'unknown')}`",
                "",
                "The live source of truth remains the executor PVC. This PR must not be reconciled back into the pod.",
            ]
        )
        return GitHubAuth._json_request(
            "POST",
            f"https://api.github.com/repos/{self.config.git_repository}/pulls",
            token=token,
            body={"title": title, "head": branch, "base": self.config.git_base_branch, "body": body},
        )

    @staticmethod
    def _git(cwd: Path, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
        env = dict(os.environ)
        env["GIT_TERMINAL_PROMPT"] = "0"
        completed = subprocess.run(
            ["git", "-c", f"safe.directory={cwd.resolve()}", *args],
            cwd=str(cwd),
            env=env,
            text=True,
            capture_output=True,
            check=False,
        )
        if check and completed.returncode != 0:
            import re
            redact_pattern = re.compile(r"x-access-token:[^@\s]+")
            stderr = redact_pattern.sub("x-access-token:***", completed.stderr)
            stdout = redact_pattern.sub("x-access-token:***", completed.stdout)
            sanitized_args = [redact_pattern.sub("x-access-token:***", arg) for arg in args]
            raise ExporterError(f"git {' '.join(sanitized_args)} failed: {stderr or stdout}")
        return completed


class SkillExportLoop:
    def __init__(self, config: ExporterConfig, git_exporter: GitSkillExporter | None = None) -> None:
        self.config = config
        self.git_exporter = git_exporter or GitSkillExporter(config)
        self.draining = threading.Event()
        self.stop_requested = threading.Event()
        self.active_lock = threading.Lock()
        self.active_cycle_id = ""
        self.last_status: dict[str, Any] = {"ok": False, "status": "starting", "started_at": _utc_now()}

    @property
    def state_path(self) -> Path:
        return self.config.state_root / "state.json"

    @property
    def checkpoint_path(self) -> Path:
        return self.config.state_root / "checkpoints/status.json"

    @property
    def ready(self) -> bool:
        return bool(self.last_status.get("ok")) and not self.draining.is_set()

    def status(self) -> dict[str, Any]:
        payload = dict(self.last_status)
        payload["drain_status"] = "draining" if self.draining.is_set() else "active"
        payload["active_cycle_id"] = self.active_cycle_id
        payload["skills_root"] = str(self.config.skills_root)
        payload["state_root"] = str(self.config.state_root)
        payload["git_repository"] = self.config.git_repository
        payload["export_root"] = self.config.export_root
        payload["pr_mode"] = self.config.pr_mode
        return payload

    def request_drain(self) -> dict[str, Any]:
        self.draining.set()
        status = self.status()
        status["ok"] = True
        status["status"] = "drain_requested"
        _atomic_json(self.checkpoint_path, status)
        return status

    def run_cycle(self) -> dict[str, Any]:
        with self.active_lock:
            if not self.config.enabled:
                status = {"ok": True, "status": "disabled", "completed_at": _utc_now()}
                self._record_status(status)
                return status
            if self.draining.is_set():
                status = {"ok": True, "status": "draining", "completed_at": _utc_now()}
                self._record_status(status)
                return status
            cycle_id = f"{int(time.time())}-{os.getpid()}"
            self.active_cycle_id = cycle_id
            try:
                try:
                    status = self._run_cycle(cycle_id)
                except Exception as exc:
                    logger.exception("Hermes skill exporter cycle failed")
                    status = {"ok": False, "status": "failed", "cycle_id": cycle_id, "error": str(exc), "completed_at": _utc_now()}
                self._record_status(status)
                return status
            finally:
                self.active_cycle_id = ""

    def _run_cycle(self, cycle_id: str) -> dict[str, Any]:
        snapshot = build_skill_snapshot(self.config.skills_root)
        state = _load_json(self.state_path, {"state_version": STATE_VERSION, "exported_tree_hashes": []})
        previous_hash = str(state.get("last_seen_tree_hash") or "")
        exported_hashes = [str(item) for item in state.get("exported_tree_hashes", []) if str(item)]
        if not state.get("baseline_tree_hash"):
            state.update(
                {
                    "state_version": STATE_VERSION,
                    "baseline_tree_hash": snapshot.tree_hash,
                    "last_seen_tree_hash": snapshot.tree_hash,
                    "baseline_recorded_at": _utc_now(),
                    "exported_tree_hashes": exported_hashes,
                }
            )
            _atomic_json(self.state_path, state)
            return self._status(cycle_id, "baseline_recorded", snapshot, previous_hash, export_result={})
        if snapshot.tree_hash == previous_hash:
            return self._status(cycle_id, "unchanged", snapshot, previous_hash, export_result={})
        if snapshot.tree_hash in exported_hashes:
            state["last_seen_tree_hash"] = snapshot.tree_hash
            _atomic_json(self.state_path, state)
            return self._status(cycle_id, "already_exported", snapshot, previous_hash, export_result={})
        metadata = self._metadata(snapshot, previous_hash)
        export_result = self.git_exporter.export(snapshot, metadata)
        if export_result.get("exported"):
            exported_hashes.append(snapshot.tree_hash)
            exported_hashes = exported_hashes[-MAX_EXPORTED_HASHES:]
            state["exported_tree_hashes"] = exported_hashes
            state["last_export"] = {
                "tree_hash": snapshot.tree_hash,
                "exported_at": _utc_now(),
                "branch": export_result.get("branch"),
                "pr_url": export_result.get("pr_url"),
                "pr_number": export_result.get("pr_number"),
            }
        state["last_seen_tree_hash"] = snapshot.tree_hash
        _atomic_json(self.state_path, state)
        return self._status(cycle_id, "exported" if export_result.get("exported") else "no_git_diff", snapshot, previous_hash, export_result=export_result)

    def _metadata(self, snapshot: SkillSnapshot, previous_hash: str) -> dict[str, Any]:
        provenance = SkillProvenanceReader(self.config.hermes_home).skill_manage_sessions()
        return {
            "schema_version": 1,
            "exported_at": _utc_now(),
            "export_environment": self.config.export_env,
            "git_repository": self.config.git_repository,
            "git_base_branch": self.config.git_base_branch,
            "pod_name": self.config.pod_name,
            "hermes_home": str(self.config.hermes_home),
            "skills_root": str(self.config.skills_root),
            "tree_hash": snapshot.tree_hash,
            "previous_tree_hash": previous_hash,
            "file_count": len(snapshot.files),
            "files": snapshot.file_manifest(),
            "skill_manage_sessions": provenance,
        }

    @staticmethod
    def _status(
        cycle_id: str,
        status: str,
        snapshot: SkillSnapshot,
        previous_hash: str,
        *,
        export_result: dict[str, Any],
    ) -> dict[str, Any]:
        return {
            "ok": True,
            "status": status,
            "cycle_id": cycle_id,
            "completed_at": _utc_now(),
            "tree_hash": snapshot.tree_hash,
            "previous_tree_hash": previous_hash,
            "file_count": len(snapshot.files),
            "export_result": export_result,
        }

    def _record_status(self, status: dict[str, Any]) -> None:
        self.last_status = status
        _atomic_json(self.checkpoint_path, status)


class ExporterServer:
    def __init__(self, config: ExporterConfig) -> None:
        self.config = config
        self.loop = SkillExportLoop(config)
        self.httpd: ThreadingHTTPServer | None = None
        self.shutdown_requested = threading.Event()

    def start(self) -> None:
        self.loop.run_cycle()
        self._start_periodic_loop()
        handler = self._handler()
        self.httpd = ThreadingHTTPServer((self.config.host, self.config.port), handler)
        self._install_signal_handlers()
        logger.info("Hermes skill exporter listening on %s:%s", self.config.host, self.config.port)
        self.httpd.serve_forever(poll_interval=0.5)

    def _install_signal_handlers(self) -> None:
        def _handle(signum, _frame) -> None:
            logger.info("Hermes skill exporter received shutdown signal %s", signum)
            self.request_shutdown_after_drain("signal")

        for signum in (signal.SIGTERM, signal.SIGINT):
            signal.signal(signum, _handle)

    def request_shutdown_after_drain(self, reason: str) -> dict[str, Any]:
        status = self.loop.request_drain()
        if self.shutdown_requested.is_set():
            return status
        self.shutdown_requested.set()

        def _shutdown() -> None:
            logger.info("Hermes skill exporter shutdown requested after drain reason=%s", reason)
            with self.loop.active_lock:
                self.loop.stop_requested.set()
            if self.httpd is not None:
                self.httpd.shutdown()

        threading.Thread(target=_shutdown, name="skill-exporter-shutdown", daemon=True).start()
        return status

    def _start_periodic_loop(self) -> None:
        def _periodic() -> None:
            while not self.loop.stop_requested.wait(self.config.sync_interval_seconds):
                self.loop.run_cycle()

        threading.Thread(target=_periodic, name="hermes-skill-exporter-loop", daemon=True).start()

    def _handler(self):
        loop = self.loop
        server = self

        class Handler(BaseHTTPRequestHandler):
            def _json(self, code: int, payload: dict[str, Any]) -> None:
                raw = json.dumps(payload, ensure_ascii=True, sort_keys=True).encode("utf-8")
                self.send_response(code)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(raw)))
                self.end_headers()
                self.wfile.write(raw)

            def do_GET(self) -> None:
                if self.path == "/healthz":
                    self._json(200, {"ok": True})
                    return
                if self.path == "/readyz":
                    self._json(200 if loop.ready else 503, loop.status())
                    return
                if self.path in {"/runtimez", "/internal/exporter/status", "/internal/drain/status"}:
                    self._json(200, loop.status())
                    return
                if self.path == "/internal/drain/start":
                    self._json(200, server.request_shutdown_after_drain("http"))
                    return
                self._json(404, {"ok": False, "error": "not found"})

            def do_POST(self) -> None:
                if self.path == "/internal/exporter/export":
                    self._json(200, loop.run_cycle())
                    return
                if self.path == "/internal/drain/start":
                    self._json(200, server.request_shutdown_after_drain("http"))
                    return
                self._json(404, {"ok": False, "error": "not found"})

            def log_message(self, fmt: str, *args: Any) -> None:
                logger.info("skill exporter http %s", fmt % args)

        return Handler


def configure_logging() -> None:
    level_name = _env("RSI_HERMES_SKILL_EXPORTER_LOG_LEVEL", _env("RSI_RUNNER_LOG_LEVEL", "INFO")).upper()
    logging.basicConfig(level=getattr(logging, level_name, logging.INFO), format="%(asctime)s %(levelname)s %(name)s %(message)s")


def main() -> None:
    parser = argparse.ArgumentParser(description="Hermes skill exporter")
    parser.parse_args()
    configure_logging()
    config = ExporterConfig.from_env()
    ExporterServer(config).start()


if __name__ == "__main__":
    main()
