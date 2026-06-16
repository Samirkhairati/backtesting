FROM python:3.11-slim AS base
WORKDIR /app
COPY ingestion/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Development stage
FROM base AS dev
# Install watchmedo for hot reloading
RUN pip install watchdog
COPY core/proto ./core/proto
COPY ingestion/ .
# Generate protos
RUN ./generate_protos.sh
# Run server with hot reload
CMD ["watchmedo", "auto-restart", "--directory=.", "--pattern=*.py", "--recursive", "--", "python", "server.py"]

# Production stage
FROM base AS prod
COPY core/proto ./core/proto
COPY ingestion/ .
RUN ./generate_protos.sh
CMD ["python", "server.py"]
