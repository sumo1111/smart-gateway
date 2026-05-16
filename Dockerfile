# ===== Build Stage: Frontend =====
FROM --platform=$BUILDPLATFORM node:20-alpine AS frontend-builder

WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install --registry=https://registry.npmmirror.com
COPY web/ ./
RUN npm run build

# ===== Build Stage: Backend =====
FROM golang:1.22-alpine AS backend-builder

RUN apk add --no-cache gcc musl-dev sqlite-dev build-base git

ENV CGO_ENABLED=1 GOOS=linux
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/build ./web/build

RUN cat VERSION && go build -trimpath -ldflags "-s -w -X 'main.Version=$(cat VERSION)' -linkmode external -extldflags '-static'" -o smart-gateway .

# ===== Runtime Stage =====
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/smart-gateway .
COPY --from=backend-builder /app/web/build ./web/build

ENV PORT=3000
ENV ADMIN_PASSWORD=admin123
ENV DB_PATH=/app/data/smart-gateway.db
ENV AUTO_STRATEGY=weighted

RUN mkdir -p /app/data
VOLUME ["/app/data"]

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:3000/api/status || exit 1

ENTRYPOINT ["./smart-gateway"]
