package main

import (
	"log"
	"net/http"

	"example.com/pz3-logging/internal/httpapi"
	"example.com/pz3-logging/internal/student"
	applogger "example.com/pz3-logging/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	logger, err := applogger.New()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	repo := student.NewRepo()
	handler := httpapi.NewHandler(repo, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/students/", handler.GetStudentByID)

	rootHandler := httpapi.LoggingMiddleware(logger, mux)

	logger.Info("server is starting",
		zap.String("addr", ":8080"),
	)

	if err := http.ListenAndServe(":8080", rootHandler); err != nil {
		logger.Fatal("server failed",
			zap.Error(err),
		)
	}
}
