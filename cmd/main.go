package main

import (
	"context"
	db2 "github.com/ahhcash/ghastlydb/db"
	"github.com/ahhcash/ghastlydb/grpc/gen/grpc/proto"
	"github.com/ahhcash/ghastlydb/grpc/server"
	http2 "github.com/ahhcash/ghastlydb/http/server"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getPort(defaultPort string) string {
	// Heroku provides PORT environment variable
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return defaultPort
}

func main() {
	// Initialize the database
	db, err := db2.OpenDB(db2.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a context that we'll use to manage shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create an error group to manage our servers
	g, ctx := errgroup.WithContext(ctx)

	// Initialize the gRPC server
	grpcServer := grpc.NewServer()
	ghastlyServer := server.NewGhastlyServer(db)
	proto.RegisterGhastlyDBServer(grpcServer, ghastlyServer)

	// Initialize the HTTP server
	httpServer := http2.NewServer(db)

	// Start gRPC server
	g.Go(func() error {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			return err
		}
		log.Printf("Starting gRPC server on :50051")
		return grpcServer.Serve(lis)
	})

	// Start HTTP server
	g.Go(func() error {
		port := getPort(":8080")
		log.Printf("Starting HTTP server on :8080")
		return httpServer.Start(port)
	})

	// Handle shutdown gracefully
	g.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigs:
			log.Printf("Received shutdown signal: %v", sig)

			// Create a timeout context for graceful shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()

			// Start graceful shutdown of HTTP server
			log.Println("Shutting down HTTP server...")
			if err := httpServer.Router().Shutdown(shutdownCtx); err != nil {
				log.Printf("HTTP server shutdown error: %v", err)
			}

			// Start graceful shutdown of gRPC server
			log.Println("Shutting down gRPC server...")
			stopped := make(chan struct{})
			go func() {
				grpcServer.GracefulStop()
				close(stopped)
			}()

			// Wait for gRPC server to stop or timeout
			select {
			case <-shutdownCtx.Done():
				log.Println("Shutdown timeout - forcing gRPC server to stop")
				grpcServer.Stop()
			case <-stopped:
				log.Println("gRPC server stopped gracefully")
			}

			// Finally cancel the main context
			cancel()
			return nil
		}
	})

	// Wait for all goroutines to complete or for an error to occur
	if err := g.Wait(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
