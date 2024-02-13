FROM golang:alpine AS build

ARG version
ARG build_time
ARG git_commit
ARG go_build_tags
ARG go_race

WORKDIR /app
RUN mkdir -p /var/run/kwil
RUN chmod 777 /var/run/kwil
RUN apk update && apk add git ca-certificates-bundle
RUN if [ -n "$go_race" ]; then apk add --no-cache gcc g++; fi

COPY . .
RUN test -f go.work && rm go.work || true

RUN GOWORK=off GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" GO_BUILDTAGS=$go_build_tags GO_RACEFLAG=$go_race ./scripts/build/binary kwild

RUN GOWORK=off GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" GO_RACEFLAG=$go_race ./scripts/build/binary kwil-admin
RUN GOWORK=off GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" GO_RACEFLAG=$go_race ./scripts/build/binary kwil-cli
RUN chmod +x /app/dist/kwild /app/dist/kwil-admin /app/dist/kwil-cli

FROM alpine:3.17
WORKDIR /app
RUN mkdir -p /var/run/kwil && chmod 777 /var/run/kwil
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/dist/kwild ./kwild
COPY --from=build /app/dist/kwil-admin ./kwil-admin
COPY --from=build /app/dist/kwil-cli ./kwil-cli
EXPOSE 50051 50151 8080 26656 26657
ENTRYPOINT ["/app/kwild"]
