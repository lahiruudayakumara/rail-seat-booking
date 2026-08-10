# Development stage with hot-reloading support
FROM node:24-alpine AS dev
RUN corepack enable
WORKDIR /src

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN pnpm install

COPY apps/web apps/web

EXPOSE 3000
CMD ["pnpm", "--filter", "@rail/web", "dev"]

# Build stage for static production assets
FROM dev AS build
ARG VITE_API_BASE_URL=http://localhost:8080
ARG VITE_PAYMENT_PROVIDER=sandbox
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_PAYMENT_PROVIDER=$VITE_PAYMENT_PROVIDER
RUN pnpm --filter @rail/web build

# Production Nginx stage
FROM nginxinc/nginx-unprivileged:1.31-alpine AS prod
COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/apps/web/dist /usr/share/nginx/html
EXPOSE 8080
