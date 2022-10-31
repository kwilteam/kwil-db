FROM golang:alpine AS stage

WORKDIR /app
RUN apk update && apk add git openssh

RUN echo -e "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN cat /root/.gitconfig
RUN mkdir /root/.ssh && echo "StrictHostKeyChecking no " > /root/.ssh/config

COPY go.mod go.sum ./
RUN --mount=type=ssh,id=kwil go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ./dist/kwild ./cmd/kwild
RUN cp *.json ./dist
RUN cp *.yaml ./dist
RUN cp *.yml ./dist
RUN cp -r ./keys ./dist/keys/
RUN cp -r ./abi ./dist/abi/

FROM scratch
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/ ./
EXPOSE 8080
EXPOSE 50051
ENTRYPOINT ["/app/kwild"]
