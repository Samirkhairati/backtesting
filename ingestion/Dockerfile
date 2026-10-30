FROM python:3.11-slim AS base
WORKDIR /app
COPY ingestion/requirements.txt .
ENV PIP_DEFAULT_TIMEOUT=100
RUN pip install --no-cache-dir -r requirements.txt

# Development stage
FROM base AS dev
# Install watchmedo for hot reloading
RUN pip install watchdog
COPY ingestion/ .
# Run server with hot reload
CMD ["watchmedo", "auto-restart", "--directory=.", "--pattern=*.py", "--recursive", "--", "python", "server.py"]

# Production stage
FROM base AS prod
COPY ingestion/ .
RUN useradd --system --no-create-home app
USER app
ENV PYTHONUNBUFFERED=1
CMD ["python", "server.py"]
