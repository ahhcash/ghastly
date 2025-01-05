package main

import (
	db2 "github.com/ahhcash/ghastlydb/db"
	pb "github.com/ahhcash/ghastlydb/grpc/gen/grpc/proto"
	"github.com/ahhcash/ghastlydb/grpc/server"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	db, err := db2.OpenDB(db2.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	grpcServer := grpc.NewServer()

	ghastlyServer := server.NewGhastlyServer(db)
	pb.RegisterGhastlyDBServer(grpcServer, ghastlyServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Starting gRPC server on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
