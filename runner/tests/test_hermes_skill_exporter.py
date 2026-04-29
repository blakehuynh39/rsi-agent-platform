from __future__ import annotations

import json
import os
from pathlib import Path
import sqlite3
import tempfile
import unittest

from rsi_runner.hermes_skill_exporter import (
    ExporterConfig,
    ExporterConfigError,
    SkillExportLoop,
    build_skill_snapshot,
    validate_export_paths,
)


class FakeGitExporter:
    def __init__(self) -> None:
        self.calls: list[tuple[object, dict]] = []

    def export(self, snapshot, metadata):  # type: ignore[no-untyped-def]
        self.calls.append((snapshot, metadata))
        return {
            "exported": True,
            "branch": f"hermes/skill-export/stage/{snapshot.tree_hash[:12]}",
            "pr_url": f"https://github.com/piplabs/rsi-agent-platform/pull/{len(self.calls)}",
            "pr_number": len(self.calls),
        }


class HermesSkillExporterTest(unittest.TestCase):
    def make_config(self, temp_dir: Path) -> ExporterConfig:
        skills_root = temp_dir / "home/skills"
        skills_root.joinpath("story-test").mkdir(parents=True)
        skills_root.joinpath("story-test/SKILL.md").write_text(
            "---\nname: story-test\ndescription: Test skill.\n---\n# Test\n",
            encoding="utf-8",
        )
        return ExporterConfig(
            host="127.0.0.1",
            port=8091,
            enabled=True,
            hermes_home=temp_dir / "home",
            skills_root=skills_root,
            state_root=temp_dir / "home/skill-exporter",
            sync_interval_seconds=300,
            git_owner="piplabs",
            git_repo="rsi-agent-platform",
            git_base_branch="main",
            export_env="stage",
            branch_prefix="hermes/skill-export",
            pr_mode="per_change",
            pod_name="test-pod",
            github_token="token",
        )

    def test_initial_baseline_records_state_without_pr(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            fake = FakeGitExporter()
            status = SkillExportLoop(config, git_exporter=fake).run_cycle()  # type: ignore[arg-type]

            self.assertEqual("baseline_recorded", status["status"])
            self.assertEqual([], fake.calls)
            state = json.loads((config.state_root / "state.json").read_text(encoding="utf-8"))
            self.assertEqual(status["tree_hash"], state["baseline_tree_hash"])

    def test_skill_tree_change_exports_one_pr_per_new_hash(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            fake = FakeGitExporter()
            loop = SkillExportLoop(config, git_exporter=fake)  # type: ignore[arg-type]
            loop.run_cycle()

            skill_md = config.skills_root / "story-test/SKILL.md"
            skill_md.write_text(skill_md.read_text(encoding="utf-8") + "\nNew lesson.\n", encoding="utf-8")
            status = loop.run_cycle()

            self.assertEqual("exported", status["status"])
            self.assertEqual(1, len(fake.calls))
            self.assertEqual(status["tree_hash"], fake.calls[0][0].tree_hash)

    def test_repeated_same_hash_is_idempotent(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            fake = FakeGitExporter()
            loop = SkillExportLoop(config, git_exporter=fake)  # type: ignore[arg-type]
            loop.run_cycle()
            skill_md = config.skills_root / "story-test/SKILL.md"
            skill_md.write_text(skill_md.read_text(encoding="utf-8") + "\nNew lesson.\n", encoding="utf-8")
            loop.run_cycle()

            status = loop.run_cycle()

            self.assertEqual("unchanged", status["status"])
            self.assertEqual(1, len(fake.calls))

    def test_exporter_refuses_state_under_skills_root(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            skills_root = root / "skills"
            skills_root.mkdir()
            with self.assertRaises(ExporterConfigError):
                ExporterConfig(
                    host="127.0.0.1",
                    port=8091,
                    enabled=True,
                    hermes_home=root,
                    skills_root=skills_root,
                    state_root=skills_root / ".state",
                    sync_interval_seconds=300,
                    git_owner="piplabs",
                    git_repo="rsi-agent-platform",
                    git_base_branch="main",
                    export_env="stage",
                    branch_prefix="hermes/skill-export",
                    pr_mode="per_change",
                    pod_name="test-pod",
                )

    def test_export_path_allowlist_blocks_non_export_paths(self) -> None:
        validate_export_paths(
            "hermes/exported-skills/stage",
            ["hermes/exported-skills/stage/skills/a/SKILL.md", "hermes/exported-skills/stage/metadata.json"],
        )
        with self.assertRaises(Exception):
            validate_export_paths("hermes/exported-skills/stage", ["runner/rsi_runner/main.py"])

    def test_metadata_includes_skill_manage_session_provenance(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            db = config.skills_root.parent / "state.db"
            connection = sqlite3.connect(db)
            connection.execute(
                "CREATE TABLE messages (session_id TEXT, timestamp REAL, role TEXT, tool_name TEXT, content TEXT, tool_calls TEXT)"
            )
            connection.execute(
                "INSERT INTO messages VALUES (?, ?, ?, ?, ?, ?)",
                ("session-1", 1.0, "assistant", "skill_manage", "{}", ""),
            )
            connection.commit()
            connection.close()
            fake = FakeGitExporter()
            loop = SkillExportLoop(config, git_exporter=fake)  # type: ignore[arg-type]
            loop.run_cycle()
            skill_md = config.skills_root / "story-test/SKILL.md"
            skill_md.write_text(skill_md.read_text(encoding="utf-8") + "\nNew lesson.\n", encoding="utf-8")

            loop.run_cycle()

            metadata = fake.calls[0][1]
            self.assertEqual("session-1", metadata["skill_manage_sessions"][0]["session_id"])
            self.assertEqual(str(config.hermes_home), metadata["hermes_home"])

    def test_provenance_uses_configured_hermes_home_when_skills_root_is_custom(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            hermes_home = root / "runtime-home"
            skills_root = root / "custom-skills"
            skills_root.joinpath("story-test").mkdir(parents=True)
            skills_root.joinpath("story-test/SKILL.md").write_text(
                "---\nname: story-test\ndescription: Test skill.\n---\n# Test\n",
                encoding="utf-8",
            )
            config = ExporterConfig(
                host="127.0.0.1",
                port=8091,
                enabled=True,
                hermes_home=hermes_home,
                skills_root=skills_root,
                state_root=hermes_home / "skill-exporter",
                sync_interval_seconds=300,
                git_owner="piplabs",
                git_repo="rsi-agent-platform",
                git_base_branch="main",
                export_env="stage",
                branch_prefix="hermes/skill-export",
                pr_mode="per_change",
                pod_name="test-pod",
                github_token="token",
            )
            hermes_home.mkdir(parents=True)
            connection = sqlite3.connect(hermes_home / "state.db")
            connection.execute(
                "CREATE TABLE messages (session_id TEXT, timestamp REAL, role TEXT, tool_name TEXT, content TEXT, tool_calls TEXT)"
            )
            connection.execute(
                "INSERT INTO messages VALUES (?, ?, ?, ?, ?, ?)",
                ("session-custom", 1.0, "assistant", "skill_manage", "{}", ""),
            )
            connection.commit()
            connection.close()
            fake = FakeGitExporter()
            loop = SkillExportLoop(config, git_exporter=fake)  # type: ignore[arg-type]
            loop.run_cycle()
            skill_md = config.skills_root / "story-test/SKILL.md"
            skill_md.write_text(skill_md.read_text(encoding="utf-8") + "\nNew lesson.\n", encoding="utf-8")

            loop.run_cycle()

            metadata = fake.calls[0][1]
            self.assertEqual("session-custom", metadata["skill_manage_sessions"][0]["session_id"])

    def test_run_cycle_records_status_while_active_lock_is_held(self) -> None:
        class ObservedLoop(SkillExportLoop):
            def __init__(self, config: ExporterConfig) -> None:
                super().__init__(config, git_exporter=FakeGitExporter())  # type: ignore[arg-type]
                self.recorded_while_locked: list[bool] = []

            def _record_status(self, status: dict) -> None:  # type: ignore[no-untyped-def]
                self.recorded_while_locked.append(self.active_lock.locked())
                super()._record_status(status)

        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            loop = ObservedLoop(config)

            loop.run_cycle()

            self.assertEqual([True], loop.recorded_while_locked)

    def test_from_env_rejects_legacy_learner_config(self) -> None:
        previous = dict(os.environ)
        try:
            os.environ.clear()
            os.environ["RSI_HERMES_LEARNER_MODE"] = "live_pvc"
            with self.assertRaises(ExporterConfigError):
                ExporterConfig.from_env()
        finally:
            os.environ.clear()
            os.environ.update(previous)

    def test_snapshot_excludes_bundled_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            first = build_skill_snapshot(config.skills_root)
            (config.skills_root / ".bundled_manifest").write_text("changed\n", encoding="utf-8")
            second = build_skill_snapshot(config.skills_root)
            self.assertEqual(first.tree_hash, second.tree_hash)


if __name__ == "__main__":
    unittest.main()
