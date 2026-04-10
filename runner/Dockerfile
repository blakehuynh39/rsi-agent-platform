FROM python:3.11-slim

WORKDIR /app
COPY runner /app
RUN pip install --no-cache-dir .

EXPOSE 8090
CMD ["python", "-m", "rsi_runner.main"]

