FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      bash ca-certificates curl git jq make openssh-client python3 python3-pip ripgrep && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
CMD ["bash", "-lc", "trap : TERM INT; sleep infinity & wait"]

