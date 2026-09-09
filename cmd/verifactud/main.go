package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	path := flag.String("config", "verifactud.json", "Path to the configuration file")

	flag.Parse()

	cfg, err := cargarConfig(*path)
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	slog.Info("Listening", "address", cfg.Listen)

	err = http.ListenAndServe(cfg.Listen, mux)

	if err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}
