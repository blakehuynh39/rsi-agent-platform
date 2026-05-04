from __future__ import annotations

from dataclasses import replace
import json
import os
from pathlib import Path
import sqlite3
import subprocess
import tempfile
import threading
import unittest
from unittest import mock

from rsi_runner.hermes_skill_exporter import (
    ExporterConfig,
    ExporterConfigError,
    ExporterServer,
    ExporterError,
    GitHubAuth,
    GitSkillExporter,
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


class FakeHTTPServer:
    def __init__(self) -> None:
        self.shutdown_called = threading.Event()

    def shutdown(self) -> None:
        self.shutdown_called.set()


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
            company_wiki_root=temp_dir / "home/wiki",
            export_wiki_enabled=True,
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

    def test_company_wiki_change_exports_markdown_under_wiki_tree(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            fake = FakeGitExporter()
            loop = SkillExportLoop(config, git_exporter=fake)  # type: ignore[arg-type]
            loop.run_cycle()
            assert config.company_wiki_root is not None
            config.company_wiki_root.joinpath("pages").mkdir(parents=True)
            config.company_wiki_root.joinpath(".staging").mkdir(parents=True)
            config.company_wiki_root.joinpath(".locks").mkdir(parents=True)
            config.company_wiki_root.joinpath("index.md").write_text("# Company Wiki Index\n", encoding="utf-8")
            config.company_wiki_root.joinpath("log.md").write_text("## [2026-05-04T00:00:00Z] ingest | Runbook\n", encoding="utf-8")
            config.company_wiki_root.joinpath("manifest.json").write_text('{"schema_version":1,"pages":{}}\n', encoding="utf-8")
            config.company_wiki_root.joinpath("pages/runbook.md").write_text("# Runbook\n", encoding="utf-8")
            config.company_wiki_root.joinpath(".staging/partial.tmp").write_text("do not export\n", encoding="utf-8")
            config.company_wiki_root.joinpath(".locks/manifest.lock").write_text("", encoding="utf-8")

            status = loop.run_cycle()

            self.assertEqual("exported", status["status"])
            self.assertEqual(1, len(fake.calls))
            snapshot = fake.calls[0][0]
            metadata = fake.calls[0][1]
            self.assertEqual(4, len(snapshot.wiki_files))
            self.assertEqual(1, len(snapshot.files))
            self.assertEqual(5, metadata["file_count"])
            self.assertEqual(4, metadata["wiki_file_count"])
            self.assertEqual(
                ["index.md", "log.md", "manifest.json", "pages/runbook.md"],
                [item.relative_path for item in snapshot.wiki_files],
            )
            self.assertTrue(metadata["wiki_root_exists"])
            self.assertTrue(metadata["wiki_tree_hash"])
            self.assertEqual(metadata["wiki_files"], snapshot.wiki_file_manifest())

            checkout = Path(raw) / "checkout"
            checkout.mkdir()
            GitSkillExporter(config)._write_export_tree(checkout, snapshot, metadata)
            export_root = checkout / config.export_root
            self.assertTrue(export_root.joinpath("skills/story-test/SKILL.md").exists())
            self.assertEqual("# Runbook\n", export_root.joinpath("wiki/pages/runbook.md").read_text(encoding="utf-8"))
            self.assertFalse(export_root.joinpath("wiki/.staging/partial.tmp").exists())
            self.assertFalse(export_root.joinpath("wiki/.locks/manifest.lock").exists())

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
            [
                "hermes/exported-skills/stage/skills/a/SKILL.md",
                "hermes/exported-skills/stage/wiki/index.md",
                "hermes/exported-skills/stage/metadata.json",
            ],
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

    def test_external_drain_stops_server_after_active_cycle(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            server = ExporterServer(config)
            fake_httpd = FakeHTTPServer()
            server.httpd = fake_httpd  # type: ignore[assignment]
            server.loop.last_status = {"ok": True, "status": "unchanged"}
            server.loop.active_lock.acquire()
            try:
                status = server.request_shutdown_after_drain("test")

                self.assertEqual("drain_requested", status["status"])
                self.assertTrue(server.loop.draining.is_set())
                self.assertFalse(server.loop.stop_requested.wait(0.05))
                self.assertFalse(fake_httpd.shutdown_called.is_set())
            finally:
                server.loop.active_lock.release()

            self.assertTrue(server.loop.stop_requested.wait(1))
            self.assertTrue(fake_httpd.shutdown_called.wait(1))

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

    def test_git_commands_allow_state_root_checkouts_with_different_owner(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            checkout = Path(raw)
            with mock.patch("rsi_runner.hermes_skill_exporter.subprocess.run") as run:
                run.return_value = subprocess.CompletedProcess(args=[], returncode=0, stdout="", stderr="")

                GitSkillExporter._git(checkout, "remote", "add", "origin", "https://example.invalid/repo.git")

            command = run.call_args.args[0]
            self.assertEqual("git", command[0])
            self.assertIn(f"safe.directory={checkout.resolve()}", command)

    def test_git_exporter_uses_isolated_checkouts_for_same_tree_hash(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            snapshot = build_skill_snapshot(config.skills_root, config.company_wiki_root)
            init_checkouts: list[Path] = []

            def fake_git(cwd: Path, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
                if args == ("init",):
                    init_checkouts.append(cwd)
                returncode = 1 if args == ("diff", "--cached", "--quiet") else 0
                return subprocess.CompletedProcess(args=list(args), returncode=returncode, stdout="", stderr="")

            with (
                mock.patch("rsi_runner.hermes_skill_exporter.GitHubAuth.token", return_value="token"),
                mock.patch.object(GitSkillExporter, "_git", side_effect=fake_git),
                mock.patch.object(
                    GitSkillExporter,
                    "_open_pr",
                    return_value={"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/42", "number": 42},
                ),
                mock.patch.object(GitSkillExporter, "_merge_pr", return_value={"merged": True}),
            ):
                GitSkillExporter(config).export(snapshot, {"pod_name": "test-pod"})
                GitSkillExporter(config).export(snapshot, {"pod_name": "test-pod"})

            self.assertEqual(2, len(init_checkouts))
            self.assertNotEqual(init_checkouts[0], init_checkouts[1])
            for checkout in init_checkouts:
                self.assertTrue(checkout.name.startswith(f"{snapshot.tree_hash[:12]}-"))
                self.assertFalse(checkout.exists())
            self.assertFalse((config.state_root / "checkouts" / snapshot.tree_hash[:12]).exists())

    def test_cycle_lock_skips_export_when_another_process_is_active(self) -> None:
        try:
            import fcntl
        except ImportError:
            self.skipTest("fcntl is unavailable on this platform")

        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            config.state_root.mkdir(parents=True)
            lock_path = config.state_root / "cycle.lock"
            lock_handle = lock_path.open("a+", encoding="utf-8")
            fcntl.flock(lock_handle.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
            try:
                fake = FakeGitExporter()
                status = SkillExportLoop(config, git_exporter=fake).run_cycle()  # type: ignore[arg-type]
            finally:
                fcntl.flock(lock_handle.fileno(), fcntl.LOCK_UN)
                lock_handle.close()

            self.assertTrue(status["ok"])
            self.assertEqual("cycle_already_running", status["status"])
            self.assertFalse((config.state_root / "state.json").exists())
            self.assertEqual([], fake.calls)

    def test_git_exporter_auto_merges_created_pr_by_default(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            snapshot = build_skill_snapshot(config.skills_root, config.company_wiki_root)

            def fake_git(_cwd: Path, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
                returncode = 1 if args == ("diff", "--cached", "--quiet") else 0
                return subprocess.CompletedProcess(args=list(args), returncode=returncode, stdout="", stderr="")

            with (
                mock.patch("rsi_runner.hermes_skill_exporter.GitHubAuth.token", return_value="token"),
                mock.patch.object(GitSkillExporter, "_git", side_effect=fake_git),
                mock.patch.object(
                    GitSkillExporter,
                    "_open_pr",
                    return_value={"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/42", "number": 42},
                ),
                mock.patch.object(GitSkillExporter, "_merge_pr", return_value={"merged": True, "sha": "abc123"}) as merge_pr,
            ):
                result = GitSkillExporter(config).export(snapshot, {"pod_name": "test-pod"})

            self.assertTrue(result["exported"])
            self.assertTrue(result["auto_merge"])
            self.assertEqual({"merged": True, "sha": "abc123"}, result["merge_result"])
            merge_pr.assert_called_once()

    def test_git_exporter_can_disable_auto_merge(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = replace(self.make_config(Path(raw)), auto_merge=False)
            snapshot = build_skill_snapshot(config.skills_root, config.company_wiki_root)

            def fake_git(_cwd: Path, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
                returncode = 1 if args == ("diff", "--cached", "--quiet") else 0
                return subprocess.CompletedProcess(args=list(args), returncode=returncode, stdout="", stderr="")

            with (
                mock.patch("rsi_runner.hermes_skill_exporter.GitHubAuth.token", return_value="token"),
                mock.patch.object(GitSkillExporter, "_git", side_effect=fake_git),
                mock.patch.object(
                    GitSkillExporter,
                    "_open_pr",
                    return_value={"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/42", "number": 42},
                ),
                mock.patch.object(GitSkillExporter, "_merge_pr") as merge_pr,
            ):
                result = GitSkillExporter(config).export(snapshot, {"pod_name": "test-pod"})

            self.assertTrue(result["exported"])
            self.assertFalse(result["auto_merge"])
            self.assertEqual({}, result["merge_result"])
            merge_pr.assert_not_called()

    def test_git_exporter_records_auto_merge_failure_without_duplicating_export(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            snapshot = build_skill_snapshot(config.skills_root, config.company_wiki_root)

            def fake_git(_cwd: Path, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
                returncode = 1 if args == ("diff", "--cached", "--quiet") else 0
                return subprocess.CompletedProcess(args=list(args), returncode=returncode, stdout="", stderr="")

            with (
                mock.patch("rsi_runner.hermes_skill_exporter.GitHubAuth.token", return_value="token"),
                mock.patch.object(GitSkillExporter, "_git", side_effect=fake_git),
                mock.patch.object(
                    GitSkillExporter,
                    "_open_pr",
                    return_value={"html_url": "https://github.com/piplabs/rsi-agent-platform/pull/42", "number": 42},
                ),
                mock.patch.object(GitSkillExporter, "_merge_pr", side_effect=ExporterError("not mergeable")),
            ):
                result = GitSkillExporter(config).export(snapshot, {"pod_name": "test-pod"})

            self.assertTrue(result["exported"])
            self.assertEqual(False, result["merge_result"]["merged"])
            self.assertIn("not mergeable", result["merge_result"]["error"])

    def test_merge_pr_uses_configured_squash_method(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            snapshot = build_skill_snapshot(config.skills_root, config.company_wiki_root)
            with mock.patch.object(GitHubAuth, "_json_request", return_value={"merged": True}) as api:
                result = GitSkillExporter(config)._merge_pr("token", 42, snapshot)

            self.assertEqual({"merged": True}, result)
            api.assert_called_once()
            args, kwargs = api.call_args
            self.assertEqual("PUT", args[0])
            self.assertEqual("https://api.github.com/repos/piplabs/rsi-agent-platform/pulls/42/merge", args[1])
            self.assertEqual("token", kwargs["token"])
            self.assertEqual("squash", kwargs["body"]["merge_method"])


if __name__ == "__main__":
    unittest.main()
