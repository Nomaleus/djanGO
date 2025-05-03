package server

import (
	"djanGO/handlers"
	pb "djanGO/proto"
	"djanGO/storage"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func StartGRPCServer(addr string, store *storage.Storage, handler *handlers.Handler) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	taskServer := NewTaskServer(store, handler)
	pb.RegisterTaskServiceServer(s, taskServer)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
