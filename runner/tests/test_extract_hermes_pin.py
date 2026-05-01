from __future__ import annotations

import importlib.util
from pathlib import Path
import tempfile
import unittest


def load_script_module():
    script_path = Path(__file__).resolve().parents[2] / "scripts/extract_hermes_pin.py"
    spec = importlib.util.spec_from_file_location("extract_hermes_pin", script_path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


class ExtractHermesPinTest(unittest.TestCase):
    def test_extracts_hermes_git_pin_from_pyproject(self) -> None:
        module = load_script_module()
        with tempfile.TemporaryDirectory() as raw:
            pyproject = Path(raw) / "pyproject.toml"
            pyproject.write_text(
                "\n".join(
                    [
                        "[project]",
                        "dependencies = [",
                        '  "requests>=2",',
                        '  "hermes-agent[honcho,mcp] @ git+https://github.com/blakehuynh39/hermes-agent.git@0123456789abcdef0123456789abcdef01234567",',
                        "]",
                    ]
                )
                + "\n",
                encoding="utf-8",
            )

            self.assertEqual(
                module.extract_hermes_pin(pyproject),
                "0123456789abcdef0123456789abcdef01234567",
            )

    def test_missing_hermes_git_pin_fails_loudly(self) -> None:
        module = load_script_module()
        with tempfile.TemporaryDirectory() as raw:
            pyproject = Path(raw) / "pyproject.toml"
            pyproject.write_text(
                "\n".join(
                    [
                        "[project]",
                        "dependencies = [",
                        '  "hermes-agent[honcho,mcp]>=0.12.0",',
                        "]",
                    ]
                )
                + "\n",
                encoding="utf-8",
            )

            with self.assertRaises(SystemExit) as raised:
                module.extract_hermes_pin(pyproject)

            self.assertIn("Failed to find pinned hermes-agent git dependency", str(raised.exception))


if __name__ == "__main__":
    unittest.main()
