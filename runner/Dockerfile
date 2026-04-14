FROM python:3.11-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends git ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY runner /app
RUN pip install --no-cache-dir .

EXPOSE 8090
CMD ["python", "-m", "rsi_runner.main"]
