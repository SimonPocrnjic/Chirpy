package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServerHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	handlerApp := http.StripPrefix("/app",http.FileServer(http.Dir(".")))
	hanlderHealth := func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	mux.Handle("/app/", middlewareMetricsInc(handlerApp))
	mux.HandleFunc("/healthz", hanlderHealth)

	var server http.Server

	server.Handler = mux
	server.Addr = ":8080"

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Println("Connection closed")
	}

}