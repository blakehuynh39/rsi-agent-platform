#!/usr/bin/env python3

from __future__ import annotations

import argparse
import re
from pathlib import Path


LINE_RE = re.compile(r"^(\s*)([A-Za-z0-9_]+):(?:\s*(.*))?$")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Update RSI image tags in story-deployments values.")
    parser.add_argument("--file", required=True, help="Path to the story-deployments values file.")
    parser.add_argument("--control-plane-tag", required=True)
    parser.add_argument("--tool-gateway-tag", required=True)
    parser.add_argument("--improvement-plane-tag", required=True)
    parser.add_argument("--runner-tag", required=True)
    parser.add_argument("--hermes-executor-tag", required=True)
    parser.add_argument("--hermes-learner-tag", required=True)
    parser.add_argument("--honcho-tag", required=True)
    parser.add_argument("--sandbox-tag", required=True)
    return parser.parse_args()


def is_mapping_start(value: str) -> bool:
    stripped = value.strip()
    return stripped == "" or stripped.startswith("&")


def update_tags(path: Path, updates: dict[tuple[str, ...], str]) -> None:
    lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    stack: list[tuple[int, str]] = []
    seen: set[tuple[str, ...]] = set()

    for index, line in enumerate(lines):
        stripped = line.lstrip()
        if stripped.startswith("- "):
            continue

        match = LINE_RE.match(line)
        if not match:
            continue

        indent = len(match.group(1))
        key = match.group(2)
        value = match.group(3) or ""

        while stack and stack[-1][0] >= indent:
            stack.pop()

        current_path = tuple([item[1] for item in stack] + [key])
        if current_path in updates:
            prefix = line[: len(line) - len(stripped)]
            lines[index] = f'{prefix}{key}: "{updates[current_path]}"\n'
            seen.add(current_path)
            value = updates[current_path]

        if is_mapping_start(value):
            stack.append((indent, key))

    missing = sorted(path_key for path_key in updates if path_key not in seen)
    if missing:
        joined = ", ".join(".".join(item) for item in missing)
        raise SystemExit(f"Failed to update expected tag paths: {joined}")

    path.write_text("".join(lines), encoding="utf-8")


def main() -> None:
    args = parse_args()
    updates: dict[tuple[str, ...], str] = {
        ("controlPlane", "image", "tag"): args.control_plane_tag,
        ("toolGateway", "image", "tag"): args.tool_gateway_tag,
        ("improvementPlane", "image", "tag"): args.improvement_plane_tag,
        ("runner", "image", "tag"): args.runner_tag,
        ("hermesExecutor", "image", "tag"): args.hermes_executor_tag,
        ("hermesLearner", "image", "tag"): args.hermes_learner_tag,
        ("honcho", "image", "tag"): args.honcho_tag,
        ("sandboxRuntime", "image", "tag"): args.sandbox_tag,
    }
    update_tags(Path(args.file), updates)


if __name__ == "__main__":
    main()
