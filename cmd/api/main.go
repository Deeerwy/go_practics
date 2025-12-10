package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	_ "net/http/pprof"
	"example.com/pprof-lab/internal/work"
)

func main() {
	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		// Добавляем таймер-декоратор
		defer work.TimeIt("Fib(38)")()

		n := 38 // достаточно тяжело для CPU
		res := work.FibFast(n)
		
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%d\n", res)
	})

	log.Println("Server on :8080; pprof on /debug/pprof/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableLocks() {
	runtime.SetBlockProfileRate(1)     // включить Block profile
	runtime.SetMutexProfileFraction(1) // включить Mutex profile
}


func fmtInt(v int) string { return fmt.Sprintf("%d\n", v) }
