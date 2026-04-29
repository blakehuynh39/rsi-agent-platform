from __future__ import annotations

import argparse
from dataclasses import dataclass, field
from datetime import datetime, timezone
import fnmatch
import hashlib
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import logging
import os
from pathlib import Path
import shutil
import signal
import tempfile
import threading
import time
from typing import Any

import yaml


logger = logging.getLogger(__name__)

LEARNER_OWNED_PATHS = (
    "knowledge/story-company/**",
    "hermes/skills/story-company/**",
    "hermes/evals/story-company/**",
    "hermes/learner/manifest.yaml",
)

CANONICAL_ROOTS = (
    "knowledge/story-company",
    "hermes/skills/story-company",
    "hermes/evals/story-company",
)
CANONICAL_FILES = ("hermes/learner/manifest.yaml",)
SYNC_STATE_VERSION = 1


class LearnerConfigError(ValueError):
    pass


class PackValidationError(ValueError):
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
    raise LearnerConfigError(f"{name} must be a boolean")


def _positive_int_env(name: str, default: int) -> int:
    raw = _env(name, str(default))
    try:
        value = int(raw)
    except ValueError as exc:
        raise LearnerConfigError(f"{name} must be a positive integer") from exc
    if value <= 0:
        raise LearnerConfigError(f"{name} must be a positive integer")
    return value


def _csv(raw: str) -> list[str]:
    seen: set[str] = set()
    out: list[str] = []
    for item in raw.split(","):
        value = item.strip()
        if not value or value in seen:
            continue
        seen.add(value)
        out.append(value)
    return out


def _utc_now() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds").replace("+00:00", "Z")


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _sha256_text(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


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
    finally:
        temp_path.unlink(missing_ok=True)


def _atomic_copy(source: Path, target: Path) -> None:
    target.parent.mkdir(parents=True, exist_ok=True)
    fd, temp_name = tempfile.mkstemp(prefix=target.name + ".", suffix=".tmp", dir=str(target.parent))
    os.close(fd)
    temp_path = Path(temp_name)
    try:
        shutil.copy2(source, temp_path)
        os.replace(temp_path, target)
    finally:
        temp_path.unlink(missing_ok=True)


def _load_json(path: Path, default: dict[str, Any] | None = None) -> dict[str, Any]:
    if not path.exists():
        return default or {}
    parsed = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(parsed, dict):
        raise ValueError(f"{path} must contain a JSON object")
    return parsed


def _safe_relative(path: str) -> str:
    value = path.replace("\\", "/").strip("/")
    if not value or value.startswith("../") or "/../" in f"/{value}/":
        raise PackValidationError(f"unsafe learner path: {path}")
    return value


def path_is_allowed(path: str, allowed_patterns: list[str]) -> bool:
    relative = _safe_relative(path)
    return any(fnmatch.fnmatch(relative, pattern) for pattern in allowed_patterns)


@dataclass(frozen=True)
class LearnerConfig:
    host: str
    port: int
    hermes_home: Path
    canonical_root: Path
    pack_path: Path
    skills_path: Path
    workspace_root: Path
    eval_output_root: Path
    ledger_root: Path
    state_root: Path
    sync_interval_seconds: int
    hermes_pin: str
    allowed_paths: list[str]
    promotion_enabled: bool
    promotion_branch_prefix: str
    git_repository: str
    git_base_branch: str
    honcho_base_url: str
    honcho_workspace: str

    @classmethod
    def from_env(cls) -> "LearnerConfig":
        hermes_home_raw = _env("HERMES_HOME")
        if not hermes_home_raw:
            raise LearnerConfigError("HERMES_HOME is required")
        hermes_home = Path(hermes_home_raw).expanduser().resolve()
        canonical_root = Path(_env("RSI_HERMES_LEARNER_CANONICAL_ROOT", "/app")).expanduser().resolve()
        workspace_root = Path(_env("RSI_HERMES_LEARNER_WORKSPACE_ROOT", "/workspace")).expanduser().resolve()
        state_root = Path(_env("RSI_HERMES_LEARNER_STATE_ROOT", str(hermes_home / "learner"))).expanduser().resolve()
        allowed_paths = _csv(_env("RSI_HERMES_LEARNER_ALLOWED_PATHS", ",".join(LEARNER_OWNED_PATHS)))
        if not allowed_paths:
            raise LearnerConfigError("RSI_HERMES_LEARNER_ALLOWED_PATHS must not be empty")
        git_owner = _env("RSI_HERMES_LEARNER_GIT_OWNER") or _env("RSI_GITHUB_OWNER", "piplabs")
        git_repo = _env("RSI_HERMES_LEARNER_GIT_REPO", "rsi-agent-platform")
        port_env = "RSI_HERMES_LEARNER_PORT" if _env("RSI_HERMES_LEARNER_PORT") else "RSI_RUNNER_PORT"
        return cls(
            host=_env("RSI_HERMES_LEARNER_HOST") or _env("RSI_RUNNER_HOST", "0.0.0.0"),
            port=_positive_int_env(port_env, 8090),
            hermes_home=hermes_home,
            canonical_root=canonical_root,
            pack_path=Path(_env("RSI_HERMES_LEARNER_PACK_PATH", str(state_root / "pack"))).expanduser().resolve(),
            skills_path=Path(_env("RSI_HERMES_LEARNER_SKILLS_PATH", str(hermes_home / "skills" / "story-company"))).expanduser().resolve(),
            workspace_root=workspace_root,
            eval_output_root=Path(_env("RSI_HERMES_LEARNER_EVAL_OUTPUT_ROOT", str(state_root / "eval-output"))).expanduser().resolve(),
            ledger_root=Path(_env("RSI_HERMES_LEARNER_LEDGER_ROOT", str(state_root / "promotion-ledger"))).expanduser().resolve(),
            state_root=state_root,
            sync_interval_seconds=_positive_int_env("RSI_HERMES_LEARNER_SYNC_INTERVAL_SECONDS", 300),
            hermes_pin=_env("RSI_HERMES_PIN"),
            allowed_paths=allowed_paths,
            promotion_enabled=_bool_env("RSI_HERMES_LEARNER_PROMOTION_ENABLED", True),
            promotion_branch_prefix=_env("RSI_HERMES_LEARNER_PROMOTION_BRANCH_PREFIX", "hermes/learning"),
            git_repository=f"{git_owner}/{git_repo}",
            git_base_branch=_env("RSI_HERMES_LEARNER_GIT_BASE_BRANCH", "main"),
            honcho_base_url=_env("RSI_HONCHO_BASE_URL"),
            honcho_workspace=_env("RSI_HONCHO_WORKSPACE"),
        )


@dataclass(frozen=True)
class PackManifest:
    pack_version: str
    hermes_pin: str
    schema_version: int
    eval_suite_version: int
    owned_paths: list[str]
    raw: dict[str, Any]


@dataclass(frozen=True)
class PromotionPolicy:
    allowed_paths: list[str]
    branch_prefix: str

    def validate_paths(self, paths: list[str]) -> None:
        invalid = [path for path in paths if not path_is_allowed(path, self.allowed_paths)]
        if invalid:
            raise PackValidationError("promotion touches non-learner paths: " + ", ".join(sorted(invalid)))

    def branch_name(self, *, now: datetime | None = None, short_id: str) -> str:
        current = now or datetime.now(timezone.utc)
        clean_id = "".join(ch for ch in short_id.lower() if ch.isalnum() or ch in {"-", "_"})[:12]
        if not clean_id:
            raise PackValidationError("promotion short_id is required")
        return f"{self.branch_prefix}/{current:%Y%m%d}-{clean_id}"


@dataclass(frozen=True)
class PackValidationResult:
    manifest: PackManifest
    eval_files: list[str]
    eval_cases: int


class PackValidator:
    def __init__(self, config: LearnerConfig) -> None:
        self.config = config
        self.policy = PromotionPolicy(config.allowed_paths, config.promotion_branch_prefix)

    def load_manifest(self) -> PackManifest:
        manifest_path = self.config.canonical_root / "hermes/learner/manifest.yaml"
        if not manifest_path.exists():
            raise PackValidationError(f"missing learner manifest: {manifest_path}")
        parsed = yaml.safe_load(manifest_path.read_text(encoding="utf-8"))
        if not isinstance(parsed, dict):
            raise PackValidationError("learner manifest must be a YAML mapping")
        missing = [key for key in ("pack_version", "hermes_pin", "schema_version", "eval_suite_version", "owned_paths") if key not in parsed]
        if missing:
            raise PackValidationError("learner manifest missing required keys: " + ", ".join(missing))
        owned_paths = parsed.get("owned_paths")
        if not isinstance(owned_paths, list) or not owned_paths or not all(isinstance(item, str) and item.strip() for item in owned_paths):
            raise PackValidationError("learner manifest owned_paths must be a non-empty string list")
        self.policy.validate_paths(owned_paths)
        hermes_pin = str(parsed["hermes_pin"]).strip()
        if self.config.hermes_pin and hermes_pin != self.config.hermes_pin:
            raise PackValidationError(
                f"learner manifest hermes_pin {hermes_pin} does not match RSI_HERMES_PIN {self.config.hermes_pin}"
            )
        return PackManifest(
            pack_version=str(parsed["pack_version"]).strip(),
            hermes_pin=hermes_pin,
            schema_version=int(parsed["schema_version"]),
            eval_suite_version=int(parsed["eval_suite_version"]),
            owned_paths=[_safe_relative(path) for path in owned_paths],
            raw=parsed,
        )

    def validate_eval_suite(self) -> tuple[list[str], int]:
        eval_root = self.config.canonical_root / "hermes/evals/story-company"
        if not eval_root.exists():
            raise PackValidationError(f"missing eval suite root: {eval_root}")
        eval_files = sorted(path for path in eval_root.rglob("*.jsonl") if path.is_file())
        if not eval_files:
            raise PackValidationError("learner eval suite must include at least one .jsonl file")
        case_count = 0
        relative_files: list[str] = []
        for eval_file in eval_files:
            relative_files.append(eval_file.relative_to(self.config.canonical_root).as_posix())
            for line_no, raw_line in enumerate(eval_file.read_text(encoding="utf-8").splitlines(), start=1):
                line = raw_line.strip()
                if not line:
                    continue
                try:
                    item = json.loads(line)
                except json.JSONDecodeError as exc:
                    raise PackValidationError(f"{eval_file}:{line_no} is not valid JSON") from exc
                if not isinstance(item, dict):
                    raise PackValidationError(f"{eval_file}:{line_no} must be a JSON object")
                for key in ("id", "prompt", "assertions", "provenance"):
                    if key not in item:
                        raise PackValidationError(f"{eval_file}:{line_no} missing {key}")
                if not isinstance(item["assertions"], list) or not item["assertions"]:
                    raise PackValidationError(f"{eval_file}:{line_no} assertions must be a non-empty list")
                provenance = item["provenance"]
                if not isinstance(provenance, dict) or not provenance.get("source") or not provenance.get("owner"):
                    raise PackValidationError(f"{eval_file}:{line_no} provenance must include source and owner")
                case_count += 1
        return relative_files, case_count

    def validate(self) -> PackValidationResult:
        manifest = self.load_manifest()
        eval_files, eval_cases = self.validate_eval_suite()
        return PackValidationResult(manifest=manifest, eval_files=eval_files, eval_cases=eval_cases)


@dataclass
class SyncResult:
    target: str
    applied: list[str] = field(default_factory=list)
    preserved_local_edits: list[str] = field(default_factory=list)
    removed: list[str] = field(default_factory=list)
    local_only: list[str] = field(default_factory=list)

    @property
    def ok(self) -> bool:
        return not self.preserved_local_edits

    def to_dict(self) -> dict[str, Any]:
        return {
            "target": self.target,
            "applied": self.applied,
            "preserved_local_edits": self.preserved_local_edits,
            "removed": self.removed,
            "local_only": self.local_only,
        }


class PackReconciler:
    def __init__(self, config: LearnerConfig) -> None:
        self.config = config
        self.state_path = config.state_root / "sync-manifest.json"

    def source_files(self) -> dict[str, Path]:
        files: dict[str, Path] = {}
        for relative_root in CANONICAL_ROOTS:
            root = self.config.canonical_root / relative_root
            if not root.exists():
                continue
            for path in sorted(root.rglob("*")):
                if path.is_file():
                    rel = path.relative_to(self.config.canonical_root).as_posix()
                    files[rel] = path
        for relative_file in CANONICAL_FILES:
            path = self.config.canonical_root / relative_file
            if path.exists():
                files[relative_file] = path
        return files

    def reconcile(self) -> dict[str, Any]:
        state = _load_json(self.state_path, {"version": SYNC_STATE_VERSION, "targets": {}})
        targets = state.setdefault("targets", {})
        all_source_files = self.source_files()
        pack_result = self._reconcile_target(
            name="pack",
            target_root=self.config.pack_path,
            source_files=all_source_files,
            state=targets.setdefault("pack", {}),
        )
        skill_sources = {
            path.removeprefix("hermes/skills/story-company/"): source
            for path, source in all_source_files.items()
            if path.startswith("hermes/skills/story-company/")
        }
        skills_result = self._reconcile_target(
            name="skills",
            target_root=self.config.skills_path,
            source_files=skill_sources,
            state=targets.setdefault("skills", {}),
        )
        state["version"] = SYNC_STATE_VERSION
        state["updated_at"] = _utc_now()
        _atomic_json(self.state_path, state)
        return {
            "ok": pack_result.ok and skills_result.ok,
            "pack": pack_result.to_dict(),
            "skills": skills_result.to_dict(),
        }

    def _reconcile_target(
        self,
        *,
        name: str,
        target_root: Path,
        source_files: dict[str, Path],
        state: dict[str, Any],
    ) -> SyncResult:
        result = SyncResult(target=name)
        tracked: dict[str, Any] = state.setdefault("files", {})
        target_root.mkdir(parents=True, exist_ok=True)
        seen = {_safe_relative(rel) for rel in source_files}

        for relative, source in source_files.items():
            relative = _safe_relative(relative)
            source_hash = _sha256_file(source)
            target = target_root / relative
            prior = tracked.get(relative, {})
            prior_target_hash = str(prior.get("target_hash", ""))
            target_hash = _sha256_file(target) if target.exists() else ""
            preexisting_untracked = bool(target_hash and relative not in tracked and target_hash != source_hash)
            if preexisting_untracked:
                result.preserved_local_edits.append(relative)
                tracked[relative] = {
                    "source_hash": source_hash,
                    "target_hash": source_hash,
                    "updated_at": _utc_now(),
                }
                continue
            local_edit = bool(target_hash and prior_target_hash and target_hash != prior_target_hash and target_hash != source_hash)
            if local_edit:
                result.preserved_local_edits.append(relative)
                continue
            if target_hash != source_hash:
                _atomic_copy(source, target)
                result.applied.append(relative)
                target_hash = source_hash
            tracked[relative] = {
                "source_hash": source_hash,
                "target_hash": target_hash,
                "updated_at": _utc_now(),
            }

        for relative in sorted(set(tracked) - seen):
            target = target_root / relative
            prior = tracked.get(relative, {})
            prior_target_hash = str(prior.get("target_hash", ""))
            target_hash = _sha256_file(target) if target.exists() else ""
            if target.exists() and target_hash and target_hash != prior_target_hash:
                result.local_only.append(relative)
                continue
            target.unlink(missing_ok=True)
            tracked.pop(relative, None)
            result.removed.append(relative)

        for target in sorted(path for path in target_root.rglob("*") if path.is_file()):
            relative = target.relative_to(target_root).as_posix()
            if relative not in tracked:
                result.local_only.append(relative)
        result.local_only = sorted(set(result.local_only))
        return result


class MigrationRunner:
    def __init__(self, config: LearnerConfig) -> None:
        self.config = config
        self.state_path = config.state_root / "migrations.json"

    def run(self, manifest: PackManifest) -> dict[str, Any]:
        desired = {
            "schema_version": manifest.schema_version,
            "eval_suite_version": manifest.eval_suite_version,
            "hermes_pin": manifest.hermes_pin,
            "pack_version": manifest.pack_version,
        }
        current = _load_json(self.state_path, {})
        if current.get("desired") == desired:
            return {"changed": False, "desired": desired}
        payload = {
            "desired": desired,
            "migrations": [
                "hermes_skill_metadata",
                "eval_result_schema",
                "derived_index_rebuild",
            ],
            "updated_at": _utc_now(),
        }
        _atomic_json(self.state_path, payload)
        return {"changed": True, "desired": desired}


class LearnerLoop:
    def __init__(self, config: LearnerConfig) -> None:
        self.config = config
        self.validator = PackValidator(config)
        self.reconciler = PackReconciler(config)
        self.migrations = MigrationRunner(config)
        self.draining = threading.Event()
        self.stop_requested = threading.Event()
        self.active_lock = threading.Lock()
        self.active_cycle_id = ""
        self.last_status: dict[str, Any] = {
            "ok": False,
            "status": "starting",
            "started_at": _utc_now(),
        }

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
        payload["hermes_home"] = str(self.config.hermes_home)
        payload["pack_path"] = str(self.config.pack_path)
        payload["skills_path"] = str(self.config.skills_path)
        payload["workspace_root"] = str(self.config.workspace_root)
        payload["eval_output_root"] = str(self.config.eval_output_root)
        payload["promotion_repository"] = self.config.git_repository
        payload["promotion_base_branch"] = self.config.git_base_branch
        return payload

    def checkpoint(self, payload: dict[str, Any]) -> None:
        _atomic_json(self.checkpoint_path, payload)

    def drain(self) -> dict[str, Any]:
        self.draining.set()
        status = self.status()
        status["status"] = "draining"
        self.checkpoint(status)
        return status

    def run_cycle(self) -> dict[str, Any]:
        if self.draining.is_set():
            status = self.status()
            status["status"] = "draining"
            status["ok"] = True
            self.checkpoint(status)
            return status

        with self.active_lock:
            cycle_id = _sha256_text(f"{time.time_ns()}:{os.getpid()}")[:12]
            self.active_cycle_id = cycle_id
            started_at = _utc_now()
            try:
                validation = self.validator.validate()
                migration_status = self.migrations.run(validation.manifest)
                sync_status = self.reconciler.reconcile()
                eval_summary = self._record_eval_summary(cycle_id, validation)
                branch = PromotionPolicy(self.config.allowed_paths, self.config.promotion_branch_prefix).branch_name(short_id=cycle_id)
                status = {
                    "ok": bool(sync_status.get("ok")),
                    "status": "synced" if sync_status.get("ok") else "local_edits_preserved",
                    "cycle_id": cycle_id,
                    "started_at": started_at,
                    "completed_at": _utc_now(),
                    "manifest": {
                        "pack_version": validation.manifest.pack_version,
                        "schema_version": validation.manifest.schema_version,
                        "eval_suite_version": validation.manifest.eval_suite_version,
                        "hermes_pin": validation.manifest.hermes_pin,
                    },
                    "eval_summary": eval_summary,
                    "migration_status": migration_status,
                    "sync_status": sync_status,
                    "promotion": {
                        "enabled": self.config.promotion_enabled,
                        "repository": self.config.git_repository,
                        "base_branch": self.config.git_base_branch,
                        "next_branch_example": branch,
                        "allowed_paths": self.config.allowed_paths,
                    },
                    "honcho": {
                        "base_url_configured": bool(self.config.honcho_base_url),
                        "workspace": self.config.honcho_workspace,
                    },
                }
            except Exception as exc:
                logger.exception("Hermes learner cycle failed")
                status = {
                    "ok": False,
                    "status": "error",
                    "cycle_id": cycle_id,
                    "started_at": started_at,
                    "completed_at": _utc_now(),
                    "error": str(exc),
                }
            self.last_status = status
            self.checkpoint(status)
            self.active_cycle_id = ""
            return status

    def _record_eval_summary(self, cycle_id: str, validation: PackValidationResult) -> dict[str, Any]:
        self.config.eval_output_root.mkdir(parents=True, exist_ok=True)
        payload = {
            "cycle_id": cycle_id,
            "recorded_at": _utc_now(),
            "mode": "schema-validation",
            "eval_files": validation.eval_files,
            "eval_cases": validation.eval_cases,
            "result": "passed",
        }
        _atomic_json(self.config.eval_output_root / f"{cycle_id}.json", payload)
        self.config.ledger_root.mkdir(parents=True, exist_ok=True)
        with (self.config.ledger_root / "promotion-ledger.jsonl").open("a", encoding="utf-8") as handle:
            handle.write(json.dumps(payload, ensure_ascii=True, sort_keys=True) + "\n")
        return payload

    def serve_forever(self) -> None:
        self.run_cycle()
        server = make_server(self.config, self)

        def _handle_shutdown(_signum, _frame) -> None:
            logger.info("Hermes learner received shutdown signal")
            self.drain()

            def _shutdown() -> None:
                with self.active_lock:
                    self.stop_requested.set()
                server.shutdown()

            threading.Thread(target=_shutdown, name="hermes-learner-shutdown", daemon=True).start()

        signal.signal(signal.SIGTERM, _handle_shutdown)
        signal.signal(signal.SIGINT, _handle_shutdown)

        def _periodic() -> None:
            while not self.stop_requested.wait(self.config.sync_interval_seconds):
                self.run_cycle()

        threading.Thread(target=_periodic, name="hermes-learner-loop", daemon=True).start()
        logger.info("Hermes learner listening on %s:%s", self.config.host, self.config.port)
        try:
            server.serve_forever()
        finally:
            server.server_close()


def make_server(config: LearnerConfig, loop: LearnerLoop) -> ThreadingHTTPServer:
    class LearnerHandler(BaseHTTPRequestHandler):
        def _json(self, status: int, payload: dict[str, Any]) -> None:
            body = json.dumps(payload, ensure_ascii=True, sort_keys=True).encode("utf-8")
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def do_GET(self) -> None:  # noqa: N802
            if self.path == "/healthz":
                self._json(200, loop.status())
                return
            if self.path == "/readyz":
                self._json(200 if loop.ready else 503, loop.status())
                return
            if self.path in {"/runtimez", "/internal/learner/status", "/internal/drain/status"}:
                self._json(200, loop.status())
                return
            self._json(404, {"error": "not found"})

        def do_POST(self) -> None:  # noqa: N802
            if self.path == "/internal/drain/start":
                self._json(202, loop.drain())
                return
            if self.path == "/internal/learner/reconcile":
                self._json(200, loop.run_cycle())
                return
            self._json(404, {"error": "not found"})

        def log_message(self, fmt: str, *args: Any) -> None:
            logger.info("learner http %s", fmt % args)

    return ThreadingHTTPServer((config.host, config.port), LearnerHandler)


def _configure_logging() -> None:
    level_name = _env("RSI_HERMES_LEARNER_LOG_LEVEL", _env("RSI_RUNNER_LOG_LEVEL", "INFO")).upper()
    logging.basicConfig(level=getattr(logging, level_name, logging.INFO), format="%(asctime)s %(levelname)s %(name)s %(message)s")


def main() -> None:
    parser = argparse.ArgumentParser(description="Durable Hermes learner")
    parser.add_argument("--once", action="store_true", help="Run one reconcile/eval cycle and exit")
    args = parser.parse_args()
    _configure_logging()
    config = LearnerConfig.from_env()
    loop = LearnerLoop(config)
    if args.once:
        status = loop.run_cycle()
        print(json.dumps(status, ensure_ascii=True, sort_keys=True))
        raise SystemExit(0 if status.get("ok") else 1)
    loop.serve_forever()


if __name__ == "__main__":
    main()
