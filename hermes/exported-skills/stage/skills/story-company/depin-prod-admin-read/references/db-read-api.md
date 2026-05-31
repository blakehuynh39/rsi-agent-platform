# Native DB Read Reference

This reference documents the Hermes-facing DB read flow. It intentionally does not contain bearer-token, pod-IP, curl, or direct control-plane request instructions.

## Agent Interface

Hermes should use the native `db_read.*` tools:

- `db_read.sources` lists available targets and caps.
- `db_read.schema` returns allowlisted schema metadata for a target.
- `db_read.validate` returns SQL repair feedback before an approval request.
- `db_read.query` creates the approval-bound DB read and pauses Hermes.
- `db_read.status` reads request state when the runtime resumes with a request reference.

`db_read.query` is a permissioned tool call. The model supplies only the target, purpose, and exact read-only SQL. Runtime-derived scope and execution auth are attached by RSI, not by the model.

## Execution Flow

1. Hermes calls `db_read.query` with exact SQL.
2. RSI creates or reuses one DB-read request for the Hermes tool call and stores a linked external-tool pause.
3. RSI validates the SQL offline and on the target side.
4. RSI posts an approval card in the original Slack thread.
5. An authorized DB-read admin approves or denies the exact SQL.
6. RSI executes approved reads through the configured stage worker or prod Lambda path.
7. RSI stores a sanitized result, updates the audit card, and resumes Hermes with the tool result.
8. Hermes writes the final user-facing answer in the original thread.

## Targets

| Target ID | Placement | Boundary | Typical Row Limit |
|-----------|-----------|----------|-------------------|
| depin-prod | prod | AWS Lambda relay | 100 |
| depin-stage | stage | stage worker | 100 |
| rsi-platform-stage | stage | stage worker | 50 |

## Approval Card

The approval card is audit UI. It should show:

- target
- requester
- exact SQL, within the Slack-safe limit
- full SQL hash
- validation attempt
- caps and expiry
- approved/denied/expired/executed status

The final explanation belongs to the resumed Hermes run, not the audit card.

## Safety Rules

- Do not bypass the native tool with terminal, Kubernetes, or hand-built network requests.
- Do not construct or forward authorization material manually.
- Do not submit multiple speculative DB reads when one query can answer the request.
- Do not self-approve. Approval is restricted to authorized DB-read admins.
- Treat sanitized DB results as the only model-visible result surface.
