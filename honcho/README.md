# Honcho Source

This directory is the source of truth for the Honcho image that RSI currently
ships as `rsi-agent-platform-honcho`.

## What We Ship Today

The build is not a vanilla upstream Honcho image. It is pinned to a reviewed
commit in the user fork:

- source repo: `https://github.com/blakehuynh39/honcho`
- source branch: `main`
- source commit: `4a711b5d69136797dfb3ef53e0738c88751c3463`
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

- Honcho runtime introspection at `/runtimez`, which `improvement-plane`
  queries directly
- DeepSeek V4 `reasoning_content` replay across OpenAI-compatible tool-call
  turns, which direct DeepSeek thinking mode requires
- DeepSeek V4 thinking-mode tool calls without the unsupported OpenAI
  `tool_choice` parameter
- a controlled pinned source of truth in the fork while RSI continues to ship
  and validate Honcho updates independently of upstream release timing

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
