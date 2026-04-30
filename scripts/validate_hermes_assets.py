#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path
import sys

from rsi_runner.hermes_skill_installer import (
    GitHubAuthConfig,
    InstallerConfig,
    SkillInstaller,
    scan_for_secret_values,
)


ROOT = Path(__file__).resolve().parents[1]


def validate_skill_pack() -> None:
    config = InstallerConfig(
        source_repo="piplabs/rsi-agent-platform",
        source_ref="local",
        source_path="hermes/skills/story-company",
        source_local_path=ROOT,
        skills_root=ROOT / ".hermes-validation/skills",
        target_relative_path="story-company",
        state_root=ROOT / ".hermes-validation/state",
        lock_path=ROOT / ".hermes-validation/.locks/skills.lock",
        dry_run=True,
        validate_only=True,
        lock_timeout_seconds=1.0,
        github_auth=GitHubAuthConfig(),
    )
    SkillInstaller(config).run()


def validate_capabilities() -> None:
    capabilities_root = ROOT / "hermes/capabilities"
    if not capabilities_root.exists():
        return
    scan_for_secret_values(capabilities_root)
    for path in sorted(capabilities_root.rglob("*.json")):
        payload = json.loads(path.read_text(encoding="utf-8"))
        if not isinstance(payload, dict):
            raise ValueError(f"{path} must contain a JSON object")
        if int(payload.get("schema_version", 0)) != 1:
            raise ValueError(f"{path} must use schema_version 1")
        capabilities = payload.get("capabilities")
        if not isinstance(capabilities, list):
            raise ValueError(f"{path} must contain a capabilities list")
        for index, capability in enumerate(capabilities):
            if not isinstance(capability, dict):
                raise ValueError(f"{path} capability {index} must be an object")
            for key in ("name", "type", "source_of_truth", "runtime_surface"):
                if not str(capability.get(key) or "").strip():
                    raise ValueError(f"{path} capability {index} missing {key}")


def main() -> int:
    validate_skill_pack()
    validate_capabilities()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
