# Base stage for dependencies
FROM node:20-alpine AS base
WORKDIR /app
RUN npm install -g pnpm

# Dev stage
FROM base AS dev
COPY package.json pnpm-lock.yaml ./
RUN pnpm install
COPY . .
CMD ["pnpm", "run", "dev", "--host"]

# Builder stage
FROM base AS builder
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build

# Prod stage
FROM nginx:1.27-alpine AS prod
COPY --from=builder /app/dist /usr/share/nginx/html
# nginx's entrypoint renders *.template files with envsubst at container start
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
ENV API_UPSTREAM=api:8080
EXPOSE 80
