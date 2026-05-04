from __future__ import annotations

import json
import sys
from typing import Any


def _json_object(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}


def _observation_field(parsed: Any, observation: dict[str, Any], key: str) -> Any:
    if hasattr(parsed, key):
        return getattr(parsed, key)
    return observation.get(key)


def _reason_for_error(message: str) -> str:
    lower = message.lower()
    if "schema" in lower:
        return "malformed_schema"
    if "requires" in lower or "required" in lower:
        return "missing_required_context"
    if "cadence" in lower or "delta" in lower:
        return "invalid_cadence_fields"
    if "redaction" in lower or "privacy" in lower:
        return "privacy_redaction_violation"
    return "unsupported_legacy_candidate"


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
    try:
        parsed = SelfReviewObservationV1.from_dict(observation)
        result = apply_turn_review_candidate(config, parsed)
        for key in (
            "gateway_session_key",
            "cadence_scope_key",
            "memory_turn_delta",
            "skill_iteration_delta",
            "skill_iteration_delta_after_last_skill_manage",
            "memory_nudge_interval",
            "skill_nudge_interval",
            "memory_tool_used",
            "skill_manage_used",
            "memory_eligible",
            "skill_eligible",
        ):
            value = _observation_field(parsed, observation, key)
            if value is not None and value != "":
                result.setdefault(key, value)
    except ValueError as exc:
        result = {
            "candidate_status": "candidate_write_failed",
            "status": "candidate_write_failed",
            "ineligible_reason": _reason_for_error(str(exc)),
            "error": str(exc)[:800],
        }
        sys.stdout.write(json.dumps(result, ensure_ascii=True, sort_keys=True) + "\n")
        sys.stdout.flush()
        return 2
    sys.stdout.write(json.dumps(result, ensure_ascii=True, sort_keys=True) + "\n")
    sys.stdout.flush()
    return 0


if __name__ == "__main__":
    raise SystemExit(_main())
