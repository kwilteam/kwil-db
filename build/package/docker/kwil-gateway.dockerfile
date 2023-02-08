FROM golang:alpine AS stage

WORKDIR /app
RUN apk update && apk add git openssh

RUN echo -e "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN mkdir /root/.ssh && echo "StrictHostKeyChecking no " > /root/.ssh/config

## COPY ./ksl/go.mod ./ksl/go.sum ./ksl/
COPY go.mod go.sum ./
RUN --mount=type=ssh,id=kwil go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ./dist/kwil-gateway ./cmd/kwil-gateway

FROM scratch
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/ ./
EXPOSE 8082
ENTRYPOINT ["/app/kwil-gateway"]
