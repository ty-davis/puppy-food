FROM node:24-bookworm-slim AS frontend-build

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-bookworm AS backend-build

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/puppy-food ./cmd/server

FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 10001 --create-home puppy \
    && mkdir -p /data \
    && chown puppy:puppy /data

WORKDIR /app
COPY --from=backend-build --chown=puppy:puppy /out/puppy-food /app/puppy-food
COPY --from=frontend-build --chown=puppy:puppy /src/frontend/dist /app/frontend/dist

ENV LISTEN_ADDRESS=:8080 \
    DATABASE_PATH=/data/puppy-food.db \
    WEB_DIST=/app/frontend/dist

VOLUME ["/data"]
EXPOSE 8080
USER puppy
ENTRYPOINT ["/app/puppy-food"]
