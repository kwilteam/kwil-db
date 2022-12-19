FROM golang:alpine AS stage

WORKDIR /app
RUN apk update && apk add git openssh

RUN echo -e "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN cat /root/.gitconfig
RUN mkdir /root/.ssh && echo "StrictHostKeyChecking no " > /root/.ssh/config

## COPY ./ksl/go.mod ./ksl/go.sum ./ksl/
COPY go.mod go.sum ./
RUN --mount=type=ssh,id=kwil go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ./dist/kwild ./cmd/kwild
COPY *.json ./dist
COPY *.yaml ./dist
COPY *.yml ./dist
COPY ./keys/ ./dist/keys/
COPY ./abi/ ./dist/abi/
COPY meta-config.yaml ./dist/meta-config.yaml
COPY manager-config.yaml ./dist/manager-config.yaml
COPY deposit-config.yaml ./dist/deposit-config.yaml

FROM scratch
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/ ./
EXPOSE 8082
EXPOSE 50051
ENTRYPOINT ["/app/kwild"]
