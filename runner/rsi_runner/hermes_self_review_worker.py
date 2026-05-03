from __future__ import annotations

import argparse
import json
import sys


def _main() -> int:
    parser = argparse.ArgumentParser(description="Run pending Hermes self-review work for a candidate.")
    parser.add_argument("--candidate-id", type=int, required=True)
    parser.add_argument("--local-only", action="store_true")
    args = parser.parse_args()

    from self_review_queue import (  # type: ignore
        SelfReviewConfig,
        run_memory_self_review,
        run_skill_self_review_batch,
    )

    config = SelfReviewConfig.from_env()
    memory = run_memory_self_review(config, args.candidate_id)
    skill = run_skill_self_review_batch(config, local_only=args.local_only)
    payload = {"ok": True, "candidate_id": args.candidate_id, "memory": memory, "skill": skill}
    sys.stdout.write(json.dumps(payload, ensure_ascii=True, sort_keys=True) + "\n")
    sys.stdout.flush()
    return 0


if __name__ == "__main__":
    raise SystemExit(_main())
