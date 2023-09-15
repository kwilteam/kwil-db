FROM golang:alpine AS stage

ARG version
ARG build_time
ARG git_commit

WORKDIR /app
RUN apk update && apk add git ca-certificates-bundle

COPY . .
RUN test -f go.work && rm go.work || true

RUN GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" ./scripts/build/binary kwild
RUN GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" ./scripts/build/binary kwil-admin
RUN chmod +x /app/dist/kwild-* /app/dist/kwil-admin-*

FROM alpine:3.17
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/kwild-* ./kwild
COPY --from=stage /app/dist/kwil-admin-* ./kwil-admin
EXPOSE 50051 50151 8080 26656 26657
ENTRYPOINT ["/app/kwild"]
