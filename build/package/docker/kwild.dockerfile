FROM golang:alpine AS stage

ARG version
ARG build_time
ARG git_commit

WORKDIR /app
RUN apk update && apk add git openssh

COPY . .
RUN rm go.work
# use `go mod vendor` to speed up build for CI & access private deps
#RUN go mod download
RUN GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time CGO_ENABLED=0 TARGET="/app/dist" ./scripts/build/binary kwild
RUN chmod +x /app/dist/kwild-*

FROM scratch
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/kwild-* ./kwild
EXPOSE 50051
EXPOSE 8080
ENTRYPOINT ["/app/kwild", "server", "start"]
