FROM golang:alpine AS stage

ARG version
ARG build_time
ARG git_commit

WORKDIR /app
RUN apk update && apk add git openssh

COPY . .
RUN test -f go.work && rm go.work || true

RUN GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" ./scripts/build/binary kwild
RUN chmod +x /app/dist/kwild-*

FROM alpine:3.17
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/kwild-* ./kwild
# Copy the startup script into the container
COPY ./startup.sh /app/startup.sh
RUN chmod +x /app/startup.sh
RUN /app/startup.sh

EXPOSE 50051 8080 26656 26657
ENTRYPOINT ["/app/kwild", "server", "start", "--config", "/app/home_dir/config.toml"]
