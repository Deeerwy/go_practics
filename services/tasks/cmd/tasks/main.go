package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"tech-ip-sem2/services/tasks/internal/client/authclient"
	taskhttp "tech-ip-sem2/services/tasks/internal/http"
	"tech-ip-sem2/services/tasks/internal/service"
)

func main() {
	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}
	grpcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	// gRPC connection to Auth.
	cc, err := grpc.Dial(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(3*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to connect to auth grpc at %s: %v", grpcAddr, err)
	}
	defer cc.Close()

	store := service.NewStore()
	authCli := authclient.New(cc)
	handler := taskhttp.NewHandler(store, authCli).Router()

	addr := ":" + port
	log.Printf("tasks service listening on %s (auth grpc=%s)", addr, grpcAddr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}


