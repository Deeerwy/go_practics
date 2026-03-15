package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	authgrpc "tech-ip-sem2/services/auth/internal/grpc"
	authhttp "tech-ip-sem2/services/auth/internal/http"
	"tech-ip-sem2/services/auth/internal/service"
)

func main() {
	httpPort := os.Getenv("AUTH_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}
	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	svc := service.NewAuthService()

	// HTTP server (login + optional HTTP verify).
	handler := authhttp.NewHandler(svc).Router()
	httpAddr := ":" + httpPort

	// gRPC server (Verify).
	grpcAddr := ":" + grpcPort
	grpcSrv := grpc.NewServer()
	authgrpcServer := authgrpc.NewServer(svc)
	authgrpc.Register(grpcSrv, authgrpcServer)

	go func() {
		log.Printf("auth HTTP service listening on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, handler); err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcAddr, err)
	}
	log.Printf("auth gRPC service listening on %s", grpcAddr)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}


