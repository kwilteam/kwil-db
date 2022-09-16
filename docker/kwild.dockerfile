FROM golang:alpine AS stage

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ./dist/kwild ./cmd/kwild
RUN cp *.json ./dist
RUN cp -r ./keys ./dist/keys/
RUN cp -r ./abi ./dist/abi/

FROM scratch
WORKDIR /app
COPY --from=stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=stage /app/dist/ ./
ENTRYPOINT ["/app/kwild"]
