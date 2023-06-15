FROM golang:alpine AS stage

# Build Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

ARG version
ARG build_time
ARG git_commit

WORKDIR /app
RUN apk update && apk add git openssh

RUN echo -e "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN cat /root/.gitconfig
RUN mkdir /root/.ssh && echo "StrictHostKeyChecking no " > /root/.ssh/config

COPY . .
# use `go mod vendor` to speed up build for CI & access private deps
#RUN go mod download
RUN GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time GO_GCFLAGS="all=-N -l" CGO_ENABLED=0 TARGET="/app/dist" ./scripts/build/binary kwild
RUN chmod +x /app/dist/kwild-*

FROM alpine:3.17

COPY --from=stage /go/bin/dlv /dlv

WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/kwild-* ./kwild

EXPOSE 40000 50051 8080

CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/app/kwild"]