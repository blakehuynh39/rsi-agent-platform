#!/usr/bin/env python3

from __future__ import annotations

import argparse
from pathlib import Path
import re


HERMES_DEP_RE = re.compile(
    r"""["']hermes-agent(?:\[[^\]]+\])?\s*@\s*git\+[^"']+@(?P<pin>[0-9a-f]{40})["']"""
)


def extract_hermes_pin(pyproject_path: Path) -> str:
    content = pyproject_path.read_text(encoding="utf-8")
    pins = {match.group("pin") for match in HERMES_DEP_RE.finditer(content)}

    if not pins:
        raise SystemExit(f"Failed to find pinned hermes-agent git dependency in {pyproject_path}")
    if len(pins) > 1:
        joined = ", ".join(sorted(pins))
        raise SystemExit(f"Found multiple hermes-agent pins in {pyproject_path}: {joined}")
    return next(iter(pins))


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Extract the pinned hermes-agent commit from runner/pyproject.toml.")
    parser.add_argument(
        "--pyproject",
        default="runner/pyproject.toml",
        help="Path to the runner pyproject.toml file.",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    print(extract_hermes_pin(Path(args.pyproject)))


if __name__ == "__main__":
    main()
