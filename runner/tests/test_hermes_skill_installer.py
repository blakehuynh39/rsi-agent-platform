from __future__ import annotations

import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest import mock

from rsi_runner.hermes_skill_installer import (
    EMPTY_TREE_HASH,
    GitHubAuthConfig,
    InstallerConfig,
    InstallerConfigError,
    InstallerError,
    SkillSourceCheckout,
    SkillInstaller,
    build_tree_snapshot,
)


class HermesSkillInstallerTest(unittest.TestCase):
    def make_pack(
        self,
        repo: Path,
        *,
        body: str = "Use Story context carefully.\n",
        expected_previous_live_tree_hash: str = EMPTY_TREE_HASH,
    ) -> tuple[Path, str]:
        source_root = repo / "hermes/skills/story-company"
        skill_dir = source_root / "story-debugging"
        skill_dir.mkdir(parents=True, exist_ok=True)
        skill_dir.joinpath("SKILL.md").write_text(
            "---\nname: story-debugging\ndescription: Debug Story systems.\n---\n" + body,
            encoding="utf-8",
        )
        target_hash = build_tree_snapshot(source_root, excluded_relative_paths={"manifest.json"}).tree_hash
        source_root.joinpath("manifest.json").write_text(
            json.dumps(
                {
                    "schema_version": 1,
                    "pack_version": "1",
                    "source_commit": "local",
                    "expected_previous_live_tree_hash": expected_previous_live_tree_hash,
                    "target_skill_tree_hash": target_hash,
                },
                ensure_ascii=True,
                indent=2,
                sort_keys=True,
            )
            + "\n",
            encoding="utf-8",
        )
        return source_root, target_hash

    def make_config(self, root: Path, repo: Path, *, dry_run: bool = False, validate_only: bool = False) -> InstallerConfig:
        return InstallerConfig(
            source_repo="piplabs/rsi-agent-platform",
            source_ref="local",
            source_path="hermes/skills/story-company",
            source_local_path=repo,
            skills_root=root / "home/skills",
            target_relative_path="story-company",
            state_root=root / "home/skill-installer",
            lock_path=root / "home/.locks/skills.lock",
            dry_run=dry_run,
            validate_only=validate_only,
            lock_timeout_seconds=1.0,
            github_auth=GitHubAuthConfig(),
        )

    def test_installs_pack_only_under_managed_skills_subtree(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            _, target_hash = self.make_pack(repo)
            config = self.make_config(root, repo)

            result = SkillInstaller(config).run()

            self.assertEqual("installed", result["status"])
            self.assertEqual(target_hash, result["target_skill_tree_hash"])
            self.assertTrue((config.skills_root / "story-company/story-debugging/SKILL.md").is_file())
            self.assertFalse((config.skills_root / "skill-installer").exists())
            state = json.loads(config.state_path.read_text(encoding="utf-8"))
            self.assertEqual(target_hash, state["last_installed_target_hash"])

    def test_repeated_install_of_same_hash_is_idempotent(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            self.make_pack(repo)
            config = self.make_config(root, repo)
            SkillInstaller(config).run()

            result = SkillInstaller(config).run()

            self.assertEqual("unchanged", result["status"])
            self.assertFalse(result["changed"])

    def test_install_staging_uses_skills_filesystem_when_state_root_differs(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            self.make_pack(repo)
            config = InstallerConfig(
                source_repo="piplabs/rsi-agent-platform",
                source_ref="local",
                source_path="hermes/skills/story-company",
                source_local_path=repo,
                skills_root=root / "home/skills",
                target_relative_path="story-company",
                state_root=root / "workspace/.rsi/skill-installer",
                lock_path=root / "home/.locks/skills.lock",
                dry_run=False,
                validate_only=False,
                lock_timeout_seconds=1.0,
                github_auth=GitHubAuthConfig(),
            )

            result = SkillInstaller(config).run()

            self.assertEqual("installed", result["status"])
            self.assertTrue(config.state_path.is_file())
            self.assertFalse((config.state_root / "staging").exists())
            self.assertFalse((config.state_root / "backups").exists())
            self.assertFalse((config.skills_root.parent / ".skill-installer-tmp").exists())

    def test_live_divergence_blocks_install_and_preserves_live_skill(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            _, first_hash = self.make_pack(repo)
            config = self.make_config(root, repo)
            SkillInstaller(config).run()
            live_skill = config.skills_root / "story-company/story-debugging/SKILL.md"
            live_skill.write_text(live_skill.read_text(encoding="utf-8") + "\nExecutor-native edit.\n", encoding="utf-8")

            shutil_repo = root / "repo-next"
            self.make_pack(shutil_repo, body="Reviewed Git change.\n", expected_previous_live_tree_hash=first_hash)
            next_config = self.make_config(root, shutil_repo)

            with self.assertRaises(InstallerError):
                SkillInstaller(next_config).run()
            self.assertIn("Executor-native edit", live_skill.read_text(encoding="utf-8"))

    def test_failed_eval_like_secret_scan_leaves_live_tree_unchanged(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            self.make_pack(repo, body="Never commit OPENROUTER_API_KEY=sk-or-v1-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
            config = self.make_config(root, repo)

            with self.assertRaises(InstallerError):
                SkillInstaller(config).run()
            self.assertFalse((config.skills_root / "story-company").exists())

    def test_validate_only_checks_manifest_without_touching_live_skills(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            self.make_pack(repo)
            config = self.make_config(root, repo, validate_only=True)

            result = SkillInstaller(config).run()

            self.assertEqual("validated", result["status"])
            self.assertFalse((config.skills_root / "story-company").exists())

    def test_target_path_must_stay_under_skills_root(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            self.make_pack(repo)
            with self.assertRaises(InstallerConfigError):
                InstallerConfig(
                    source_repo="piplabs/rsi-agent-platform",
                    source_ref="local",
                    source_path="hermes/skills/story-company",
                    source_local_path=repo,
                    skills_root=root / "home/skills",
                    target_relative_path="../escape",
                    state_root=root / "home/skill-installer",
                    lock_path=root / "home/.locks/skills.lock",
                    dry_run=False,
                    validate_only=False,
                    lock_timeout_seconds=1.0,
                    github_auth=GitHubAuthConfig(),
                )

    def test_checkout_git_commands_allow_state_root_checkouts_with_different_owner(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            checkout = Path(raw)
            with mock.patch("rsi_runner.hermes_skill_installer.subprocess.run") as run:
                run.return_value = subprocess.CompletedProcess(args=[], returncode=0, stdout="", stderr="")

                SkillSourceCheckout._git(checkout, "remote", "add", "origin", "https://example.invalid/repo.git")

            command = run.call_args.args[0]
            self.assertEqual("git", command[0])
            self.assertIn(f"safe.directory={checkout.resolve()}", command)

    def test_checkout_fetches_branch_heads_when_private_sha_fetch_is_hidden(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            source_ref = "98f04e309dfd5e73fa461b1f844cb7117f861f0e"
            config = InstallerConfig(
                source_repo="piplabs/rsi-agent-platform",
                source_ref=source_ref,
                source_path="hermes/skills/story-company",
                source_local_path=None,
                skills_root=root / "home/skills",
                target_relative_path="story-company",
                state_root=root / "state",
                lock_path=root / "home/.locks/skills.lock",
                dry_run=False,
                validate_only=False,
                lock_timeout_seconds=1.0,
                github_auth=GitHubAuthConfig(github_token="token"),
            )
            calls: list[tuple[str, ...]] = []

            def fake_git(cwd: Path, *args: str) -> None:
                calls.append(args)
                if args == ("fetch", "--depth=1", "origin", source_ref):
                    raise InstallerError("git fetch failed: not our ref")
                if args == ("checkout", "--detach", source_ref):
                    cwd.joinpath("hermes/skills/story-company").mkdir(parents=True, exist_ok=True)

            with mock.patch.object(SkillSourceCheckout, "_git", side_effect=fake_git):
                with SkillSourceCheckout(config) as source_root:
                    self.assertTrue(source_root.is_dir())

            self.assertIn(("fetch", "--depth=200", "origin", "+refs/heads/*:refs/remotes/origin/*"), calls)
            self.assertIn(("checkout", "--detach", source_ref), calls)


if __name__ == "__main__":
    unittest.main()
