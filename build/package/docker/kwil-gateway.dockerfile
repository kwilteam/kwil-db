FROM golang:alpine AS stage

ARG version
ARG build_time
ARG git_commit
ARG git_version_commit=unknown

WORKDIR /app
RUN apk update && apk add git openssh

RUN echo -e "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN mkdir /root/.ssh && echo "StrictHostKeyChecking no " > /root/.ssh/config

COPY . .
RUN go mod download
RUN GIT_VERSION=$version GIT_COMMIT=$git_commit BUILD_TIME=$build_time GIT_VERSION_COMMIT=$git_version_commit CGO_ENABLED=0 TARGET="/app/dist" ./scripts/build/binary kwil-gateway
RUN chmod +x /app/dist/kwil-gateway-*

FROM scratch
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/kwil-gateway-* ./kwil-gateway
EXPOSE 8082
ENTRYPOINT ["/app/kwil-gateway"]
