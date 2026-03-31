package main

import (
	"log"
	"net/http"

	"example.com/pz4-monitoring/internal/httpapi"
	"example.com/pz4-monitoring/internal/student"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	repo := student.NewRepo()
	handler := httpapi.NewHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/students/", handler.GetStudentByID)
	mux.Handle("/metrics", promhttp.Handler())

	rootHandler := httpapi.MetricsMiddleware(mux)

	log.Println("server started on :8080")

	if err := http.ListenAndServe(":8080", rootHandler); err != nil {
		log.Fatal(err)
	}
}
