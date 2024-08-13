FROM golang:1.23 AS build

ARG version
ARG build_time
ARG git_commit
ARG go_build_tags
ARG go_race

WORKDIR /app
RUN mkdir -p /var/run/kwil
RUN chmod 777 /var/run/kwil

COPY . .
RUN test -f go.work && rm go.work || true

RUN GOWORK=off GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" GO_BUILDTAGS=$go_build_tags GO_RACEFLAG=$go_race ./scripts/build/binary kwild

RUN GOWORK=off GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" GO_RACEFLAG=$go_race ./scripts/build/binary kwil-admin
RUN GOWORK=off GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" GO_RACEFLAG=$go_race ./scripts/build/binary kwil-cli
RUN chmod +x /app/dist/kwild /app/dist/kwil-admin /app/dist/kwil-cli

FROM ubuntu:24.04
WORKDIR /app
RUN mkdir -p /var/run/kwil && chmod 777 /var/run/kwil 
RUN apt update &&  apt install -y postgresql-client
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/dist/kwild ./kwild
COPY --from=build /app/dist/kwil-admin ./kwil-admin
COPY --from=build /app/dist/kwil-cli ./kwil-cli
EXPOSE 8080 8484 26656 26657
ENTRYPOINT ["/app/kwild"]
