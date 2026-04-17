# Honcho Source

This directory is the source of truth for the Honcho image that RSI currently
ships as `rsi-agent-platform-honcho`.

## What We Ship Today

The build is not a vanilla upstream Honcho image. It is pinned to a reviewed
commit in the user fork:

- source repo: `https://github.com/blakehuynh39/honcho`
- source branch: `codex/first-principles-agentic-workflow-hardening`
- source commit: `832c08f884f7f4d62ee196f2f377883c8fc5011e`
- upstream base repo: `https://github.com/plastic-labs/honcho`

The current build path is:

1. Clone the pinned fork branch from [`honcho/Dockerfile`](./Dockerfile).
2. Check out the pinned fork commit.
3. Build and push `rsi-agent-platform-honcho:honcho-<sha>`.

That means stage is currently running a custom Honcho fork, not upstream Honcho
unchanged.

## Why The Fork Exists

The fork exists because RSI currently needs Honcho behavior that is not yet
available from a piplabs-owned fork or an upstream release, including:

- GPT-5 tool and reasoning requests routed through the OpenAI Responses API
- Honcho runtime settings that expose and honor RSI's configured
  `reasoning_effort`
- Honcho runtime introspection at `/runtimez`

If the pinned fork commit changes, update this document in the same change.

## How To Verify A Running Image

The image carries labels that identify the fork source and upstream base:

- `io.storyprotocol.rsi.honcho.source_repo`
- `io.storyprotocol.rsi.honcho.source_branch`
- `io.storyprotocol.rsi.honcho.source_commit`
- `io.storyprotocol.rsi.honcho.upstream_repo`
- `io.storyprotocol.rsi.honcho.fork_reason`

That lets operators confirm from ECR or `docker inspect` that the deployed image
is a custom RSI Honcho fork build.

## Intended End State

The cleaner long-term shape is a dedicated `piplabs/honcho` fork that becomes
the maintained source of truth instead of a user fork.

When that repo exists, this directory should be simplified to:

1. Clone `piplabs/honcho` at a pinned commit.
2. Update the pinned source labels and docs to reference that repo.
3. Keep this README pointing at the actual build source commit.
