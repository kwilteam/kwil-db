package main

import (
	"fmt"
	"kwil/x/grpcx"
	"kwil/x/logx"
	schemapb "kwil/x/proto/schemasvc"
	"kwil/x/svc/schemasvc"
	"net"
	"os"
)

func run(logger logx.Logger) error {
	server := grpcx.NewServer(logger)

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	srv := schemasvc.New()
	schemapb.RegisterSchemaServiceServer(server, srv)
	return server.Serve(listener)
}

func main() {
	logger := logx.New()
	if err := run(logger); err != nil {
		logger.Sugar().Error(err)
		os.Exit(-1)
	}
}
