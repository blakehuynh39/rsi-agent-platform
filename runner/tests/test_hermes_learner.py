from __future__ import annotations

from datetime import datetime, timezone
import json
import os
from pathlib import Path
import tempfile
import unittest

from rsi_runner.hermes_learner import (
    LEARNER_OWNED_PATHS,
    LearnerConfig,
    LearnerLoop,
    MigrationRunner,
    PackReconciler,
    PackValidationError,
    PackValidator,
    PromotionPolicy,
)


HERMES_PIN = "3e61703b08f475c9982cf4099d049eeac232a7a1"


def write_pack(root: Path) -> None:
    root.joinpath("knowledge/story-company").mkdir(parents=True, exist_ok=True)
    root.joinpath("knowledge/story-company/README.md").write_text("Story knowledge\n", encoding="utf-8")
    root.joinpath("hermes/skills/story-company/architecture").mkdir(parents=True, exist_ok=True)
    root.joinpath("hermes/skills/story-company/architecture/SKILL.md").write_text(
        "---\nname: architecture\n---\n# Architecture\n",
        encoding="utf-8",
    )
    root.joinpath("hermes/evals/story-company").mkdir(parents=True, exist_ok=True)
    root.joinpath("hermes/evals/story-company/smoke.jsonl").write_text(
        json.dumps(
            {
                "id": "smoke",
                "prompt": "Use story-deployments as source of truth.",
                "assertions": [{"type": "contains", "value": "story-deployments"}],
                "provenance": {"source": "test", "owner": "rsi-agent-platform"},
            }
        )
        + "\n",
        encoding="utf-8",
    )
    root.joinpath("hermes/learner").mkdir(parents=True, exist_ok=True)
    root.joinpath("hermes/learner/manifest.yaml").write_text(
        "\n".join(
            [
                'pack_version: "0.1.0"',
                "schema_version: 1",
                "eval_suite_version: 1",
                f'hermes_pin: "{HERMES_PIN}"',
                "owned_paths:",
                "  - knowledge/story-company/**",
                "  - hermes/skills/story-company/**",
                "  - hermes/evals/story-company/**",
                "  - hermes/learner/manifest.yaml",
            ]
        )
        + "\n",
        encoding="utf-8",
    )


class HermesLearnerTest(unittest.TestCase):
    def make_config(self, temp_dir: Path) -> LearnerConfig:
        canonical = temp_dir / "repo"
        write_pack(canonical)
        return LearnerConfig(
            host="127.0.0.1",
            port=8090,
            hermes_home=temp_dir / "home",
            canonical_root=canonical,
            pack_path=temp_dir / "home/learner/pack",
            skills_path=temp_dir / "home/skills/story-company",
            workspace_root=temp_dir / "workspace",
            eval_output_root=temp_dir / "home/learner/eval-output",
            ledger_root=temp_dir / "home/learner/promotion-ledger",
            state_root=temp_dir / "home/learner",
            sync_interval_seconds=300,
            hermes_pin=HERMES_PIN,
            allowed_paths=list(LEARNER_OWNED_PATHS),
            promotion_enabled=True,
            promotion_branch_prefix="hermes/learning",
            git_repository="piplabs/rsi-agent-platform",
            git_base_branch="main",
            honcho_base_url="http://honcho:8000",
            honcho_workspace="rsi-stage",
        )

    def test_reconcile_preserves_local_unpromoted_experiments(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            reconciler = PackReconciler(config)
            first = reconciler.reconcile()
            self.assertTrue(first["ok"])
            local_file = config.pack_path / "knowledge/story-company/README.md"
            local_file.write_text("local unpromoted edit\n", encoding="utf-8")

            source_file = config.canonical_root / "knowledge/story-company/README.md"
            source_file.write_text("new canonical edit\n", encoding="utf-8")
            second = reconciler.reconcile()

            self.assertFalse(second["ok"])
            self.assertIn("knowledge/story-company/README.md", second["pack"]["preserved_local_edits"])
            self.assertEqual("local unpromoted edit\n", local_file.read_text(encoding="utf-8"))

    def test_reconcile_keeps_preexisting_untracked_file_across_cycles(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            preexisting = config.pack_path / "knowledge/story-company/README.md"
            preexisting.parent.mkdir(parents=True, exist_ok=True)
            preexisting.write_text("preexisting local draft\n", encoding="utf-8")
            reconciler = PackReconciler(config)

            first = reconciler.reconcile()
            second = reconciler.reconcile()

            self.assertFalse(first["ok"])
            self.assertFalse(second["ok"])
            self.assertIn("knowledge/story-company/README.md", first["pack"]["preserved_local_edits"])
            self.assertIn("knowledge/story-company/README.md", second["pack"]["preserved_local_edits"])
            self.assertEqual("preexisting local draft\n", preexisting.read_text(encoding="utf-8"))

    def test_pack_manifest_detects_unsafe_owned_path(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            manifest = config.canonical_root / "hermes/learner/manifest.yaml"
            manifest.write_text(
                "\n".join(
                    [
                        'pack_version: "0.1.0"',
                        "schema_version: 1",
                        "eval_suite_version: 1",
                        f'hermes_pin: "{HERMES_PIN}"',
                        "owned_paths:",
                        "  - internal/runtime/**",
                    ]
                )
                + "\n",
                encoding="utf-8",
            )
            with self.assertRaises(PackValidationError):
                PackValidator(config).validate()

    def test_promotion_allowlist_blocks_unsafe_paths(self) -> None:
        policy = PromotionPolicy(list(LEARNER_OWNED_PATHS), "hermes/learning")
        policy.validate_paths(["knowledge/story-company/README.md", "hermes/learner/manifest.yaml"])
        with self.assertRaises(PackValidationError):
            policy.validate_paths(["runner/rsi_runner/main.py"])
        branch = policy.branch_name(now=datetime(2026, 4, 29, tzinfo=timezone.utc), short_id="ABC123!")
        self.assertEqual("hermes/learning/20260429-abc123", branch)

    def test_schema_migration_is_idempotent(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            manifest = PackValidator(config).load_manifest()
            runner = MigrationRunner(config)
            first = runner.run(manifest)
            second = runner.run(manifest)
            self.assertTrue(first["changed"])
            self.assertFalse(second["changed"])

    def test_run_cycle_records_checkpoint_and_eval_output(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            config = self.make_config(Path(raw))
            status = LearnerLoop(config).run_cycle()
            self.assertTrue(status["ok"])
            self.assertTrue((config.state_root / "checkpoints/status.json").exists())
            eval_files = list(config.eval_output_root.glob("*.json"))
            self.assertEqual(1, len(eval_files))
            checkpoint = json.loads((config.state_root / "checkpoints/status.json").read_text(encoding="utf-8"))
            self.assertEqual("synced", checkpoint["status"])

    def test_from_env_uses_learner_defaults(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            previous = dict(os.environ)
            try:
                os.environ.clear()
                os.environ.update(
                    {
                        "HERMES_HOME": str(Path(raw) / "home"),
                        "RSI_HERMES_PIN": HERMES_PIN,
                        "RSI_HONCHO_WORKSPACE": "rsi-stage",
                    }
                )
                config = LearnerConfig.from_env()
                self.assertEqual((Path(raw) / "home").resolve(), config.hermes_home)
                self.assertEqual(8090, config.port)
                self.assertEqual("piplabs/rsi-agent-platform", config.git_repository)
            finally:
                os.environ.clear()
                os.environ.update(previous)


if __name__ == "__main__":
    unittest.main()
