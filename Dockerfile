# Base stage for shared dependencies
FROM golang:1.26-alpine AS base
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY . .
RUN go mod tidy

# Dev stage for hot reloading
FROM base AS dev
RUN go install github.com/air-verse/air@v1.65.3
# Air will watch the /app directory and rebuild/run on file changes
CMD sh -c "go mod tidy && air"

# Builder stage for compiling the production binary
FROM base AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker

# Production stage
FROM alpine:latest AS prod
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/worker .
CMD ["./worker"]
