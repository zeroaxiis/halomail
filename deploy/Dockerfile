# syntax=docker/dockerfile:1
#
# Builds any HaloMail Go service (or the monolith) from the monorepo into a
# runtime image with PostgreSQL backup/restore tools.
#
#   docker build -f deploy/Dockerfile --build-arg SERVICE=gateway -t halomail-api .
#
# Requires generated code to be present (run `task proto` and commit, or
# generate in CI before building).

ARG GO_VERSION=1.25
ARG POSTGRES_VERSION=18

# ---- build ----
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates

# Which service under services/<SERVICE> to build. Each exposes cmd/server.
ARG SERVICE=gateway

# Build via each module's go.mod + replace directives, independent of the
# workspace (matches CI; avoids go.work listing not-yet-present modules).
ENV GOWORK=off

# All modules: the target service imports the shared module locally.
COPY go.work ./
COPY services ./services

WORKDIR /src/services/${SERVICE}
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/app ./cmd/server
RUN cd /src/services/shared && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/backup ./cmd/backup

# ---- runtime ----
FROM postgres:${POSTGRES_VERSION}-bookworm
COPY --from=build /out/app /app
COPY --from=build /out/backup /backup
EXPOSE 8080
USER postgres:postgres
ENTRYPOINT ["/app"]
CMD []
