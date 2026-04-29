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
                },
            )
            rendered = values.read_text(encoding="utf-8")
            self.assertIn('tag: "runner-new"', rendered)
            self.assertIn('tag: "hermes-executor-new"', rendered)
            self.assertIn('tag: "hermes-skill-exporter-new"', rendered)

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
                    },
                )
            self.assertIn("hermesSkillExporter.image.tag", str(raised.exception))


if __name__ == "__main__":
    unittest.main()
