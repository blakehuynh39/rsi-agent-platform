FROM python:3.11-slim

ARG HERMES_AGENT_REF=0e336b0e717027cbb81fcb5816246b7aec2d4a47

RUN apt-get update && \
    apt-get install -y --no-install-recommends git ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY runner /app
RUN pip install --no-cache-dir "git+https://github.com/NousResearch/hermes-agent.git@${HERMES_AGENT_REF}" && \
    pip install --no-cache-dir .

EXPOSE 8090
CMD ["python", "-m", "rsi_runner.main"]
