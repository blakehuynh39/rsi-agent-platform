from __future__ import annotations

import json
import sys
from typing import Any


def _json_object(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def _main() -> int:
    if len(sys.argv) < 2:
        raise RuntimeError("candidate writer input path is required")
    payload = json.loads(open(sys.argv[1], "r", encoding="utf-8").read())
    if not isinstance(payload, dict):
        raise ValueError("candidate writer input must be a JSON object")
    observation = _json_object(payload.get("observation"))
    if not observation:
        raise ValueError("self-review observation is required")

    from self_review_contracts import SelfReviewObservationV1  # type: ignore
    from self_review_queue import SelfReviewConfig, apply_turn_review_candidate  # type: ignore

    config = SelfReviewConfig.from_env()
    result = apply_turn_review_candidate(config, SelfReviewObservationV1.from_dict(observation))
    sys.stdout.write(json.dumps(result, ensure_ascii=True, sort_keys=True) + "\n")
    sys.stdout.flush()
    return 0


if __name__ == "__main__":
    raise SystemExit(_main())
