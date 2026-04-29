FROM python:3.11-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends git gh ca-certificates librsvg2-bin && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY runner /app
RUN HERMES_PIN="$(python3 -c 'import pathlib,re; text=pathlib.Path("/app/pyproject.toml").read_text(encoding="utf-8"); match=re.search(r"hermes-agent\[honcho,mcp\]\s*@\s*git\+https://github\.com/NousResearch/hermes-agent\.git@([0-9a-f]+)", text); assert match, "failed to parse Hermes pin from pyproject.toml"; print(match.group(1))')" && \
    git clone --filter=blob:none https://github.com/NousResearch/hermes-agent.git /tmp/hermes-agent && \
    git -C /tmp/hermes-agent checkout "$HERMES_PIN" && \
    mkdir -p /opt && \
    cp -R /tmp/hermes-agent/skills /opt/hermes-bundled-skills && \
    rm -rf /tmp/hermes-agent && \
    pip install --no-cache-dir /app

ENV HERMES_BUNDLED_SKILLS=/opt/hermes-bundled-skills

EXPOSE 8090
CMD ["python", "-m", "rsi_runner.main"]
