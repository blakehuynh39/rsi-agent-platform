from __future__ import annotations

import importlib.util
from pathlib import Path
import tempfile
import unittest


def load_script_module():
    script_path = Path(__file__).resolve().parents[2] / "scripts/update_story_deployments_tags.py"
    spec = importlib.util.spec_from_file_location("update_story_deployments_tags", script_path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


class UpdateStoryDeploymentsTagsTest(unittest.TestCase):
    def test_updates_hermes_skill_exporter_independently(self) -> None:
        module = load_script_module()
        with tempfile.TemporaryDirectory() as raw:
            values = Path(raw) / "values.yaml"
            values.write_text(
                "\n".join(
                    [
                        "runner:",
                        "  image:",
                        '    tag: "runner-old"',
                        "hermesExecutor:",
                        "  image:",
                        '    tag: "hermes-executor-old"',
                        "hermesSkillExporter:",
                        "  image:",
                        '    tag: "hermes-skill-exporter-old"',
                        "hermesSkillInstaller:",
                        "  enabled: false",
                        "  image:",
                        '    tag: "hermes-skill-exporter-old"',
                        "  config:",
                        '    RSI_HERMES_SKILL_INSTALLER_SOURCE_REF: "old-ref"',
                    ]
                )
                + "\n",
                encoding="utf-8",
            )
            module.update_tags(
                values,
                {
                    ("runner", "image", "tag"): "runner-new",
                    ("hermesExecutor", "image", "tag"): "hermes-executor-new",
                    ("hermesSkillExporter", "image", "tag"): "hermes-skill-exporter-new",
                    ("hermesSkillInstaller", "enabled"): True,
                    ("hermesSkillInstaller", "image", "tag"): "hermes-skill-exporter-new",
                    ("hermesSkillInstaller", "config", "RSI_HERMES_SKILL_INSTALLER_SOURCE_REF"): "new-ref",
                },
            )
            rendered = values.read_text(encoding="utf-8")
            self.assertIn('tag: "runner-new"', rendered)
            self.assertIn('tag: "hermes-executor-new"', rendered)
            self.assertIn('tag: "hermes-skill-exporter-new"', rendered)
            self.assertIn("enabled: true", rendered)
            self.assertNotIn('tag: "hermes-skill-installer-new"', rendered)
            self.assertIn('RSI_HERMES_SKILL_INSTALLER_SOURCE_REF: "new-ref"', rendered)

    def test_updates_hermes_skill_installer_source_ref_in_common_anchor(self) -> None:
        module = load_script_module()
        with tempfile.TemporaryDirectory() as raw:
            values = Path(raw) / "values.yaml"
            values.write_text(
                "\n".join(
                    [
                        "hermesSkillInstallerCommonConfig: &hermesSkillInstallerCommonConfig",
                        '  RSI_HERMES_SKILL_INSTALLER_SOURCE_REF: "old-ref"',
                        "hermesSkillInstaller:",
                        "  enabled: false",
                        "  image:",
                        '    tag: "hermes-skill-exporter-old"',
                        "  config:",
                        "    <<: *hermesSkillInstallerCommonConfig",
                    ]
                )
                + "\n",
                encoding="utf-8",
            )
            module.update_tags(
                values,
                {
                    ("hermesSkillInstaller", "enabled"): True,
                    ("hermesSkillInstaller", "image", "tag"): "hermes-skill-exporter-new",
                },
                {
                    (
                        ("hermesSkillInstaller", "config", "RSI_HERMES_SKILL_INSTALLER_SOURCE_REF"),
                        ("hermesSkillInstallerCommonConfig", "RSI_HERMES_SKILL_INSTALLER_SOURCE_REF"),
                    ): "new-ref",
                },
            )

            rendered = values.read_text(encoding="utf-8")
            self.assertIn("enabled: true", rendered)
            self.assertIn('tag: "hermes-skill-exporter-new"', rendered)
            self.assertIn('RSI_HERMES_SKILL_INSTALLER_SOURCE_REF: "new-ref"', rendered)

    def test_missing_hermes_skill_installer_source_ref_alternatives_fail_loudly(self) -> None:
        module = load_script_module()
        with tempfile.TemporaryDirectory() as raw:
            values = Path(raw) / "values.yaml"
            values.write_text(
                "\n".join(
                    [
                        "hermesSkillInstaller:",
                        "  enabled: false",
                    ]
                )
                + "\n",
                encoding="utf-8",
            )
            with self.assertRaises(SystemExit) as raised:
                module.update_tags(
                    values,
                    {
                        ("hermesSkillInstaller", "enabled"): True,
                    },
                    {
                        (
                            ("hermesSkillInstaller", "config", "RSI_HERMES_SKILL_INSTALLER_SOURCE_REF"),
                            ("hermesSkillInstallerCommonConfig", "RSI_HERMES_SKILL_INSTALLER_SOURCE_REF"),
                        ): "new-ref",
                    },
                )

            message = str(raised.exception)
            self.assertIn("hermesSkillInstaller.config.RSI_HERMES_SKILL_INSTALLER_SOURCE_REF", message)
            self.assertIn("hermesSkillInstallerCommonConfig.RSI_HERMES_SKILL_INSTALLER_SOURCE_REF", message)

    def test_missing_hermes_skill_exporter_path_fails_loudly(self) -> None:
        module = load_script_module()
        with tempfile.TemporaryDirectory() as raw:
            values = Path(raw) / "values.yaml"
            values.write_text(
                "\n".join(
                    [
                        "runner:",
                        "  image:",
                        '    tag: "runner-old"',
                    ]
                )
                + "\n",
                encoding="utf-8",
            )
            with self.assertRaises(SystemExit) as raised:
                module.update_tags(
                    values,
                    {
                        ("runner", "image", "tag"): "runner-new",
                        ("hermesSkillExporter", "image", "tag"): "hermes-skill-exporter-new",
                        ("hermesSkillInstaller", "image", "tag"): "hermes-skill-installer-new",
                    },
                )
            message = str(raised.exception)
            self.assertIn("hermesSkillExporter.image.tag", message)
            self.assertIn("hermesSkillInstaller.image.tag", message)


if __name__ == "__main__":
    unittest.main()
