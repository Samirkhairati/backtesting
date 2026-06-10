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
FROM nginx:alpine AS prod
COPY --from=builder /app/dist /usr/share/nginx/html
# Copy custom nginx config if needed, otherwise default is usually fine for static files
# To support React Router history API, we usually need a custom nginx.conf
RUN echo 'server { \
    listen 80; \
    location / { \
        root /usr/share/nginx/html; \
        index index.html index.htm; \
        try_files $uri $uri/ /index.html; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
