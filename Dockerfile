FROM node:20.19.1-bookworm AS front
WORKDIR /app
RUN npm install -g pnpm@10.10.0

# Cache pnpm store between builds
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm config set store-dir /root/.local/share/pnpm/store

COPY front/package.json front/pnpm-lock.yaml front/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

COPY front/. .
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm run generate

FROM golang:1.23.3-alpine AS backend
ARG VERSION
ARG COMMIT_ID
WORKDIR /app
RUN apk add --no-cache build-base tzdata

# Cache Go module downloads and build cache
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download

COPY backend/. .
COPY --from=front /app/.output/public /app/public
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -tags prod -ldflags="-s -w -X main.version=${VERSION} -X main.commitId=${COMMIT_ID}" -o /app/moments

FROM alpine
WORKDIR /app/data
RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV PORT=3000
ENV TZ=Asia/Shanghai
COPY --from=backend /app/moments /app/moments
RUN chmod +x /app/moments
EXPOSE 3000
CMD ["/app/moments"]
