package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	ovn "github.com/mstinsky/ovn-exporter/ovnmonitor"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := ovn.ParseFlags()
	if err != nil {
		slog.Error("failed to parse config", "error", err)
		os.Exit(1)
	}

	exporter := ovn.NewExporter(config)
	if err = exporter.StartConnection(); err != nil {
		slog.Error("failed to connect db socket", "error", err)
		go exporter.TryClientConnection()
	}
	exporter.StartOvnMetrics()
	mux := http.NewServeMux()
	if config.EnableMetrics {
		mux.Handle(config.MetricsPath, promhttp.Handler())
		slog.Info(fmt.Sprintf("Listening on %s", config.ListenAddress))
	}

	// conform to Gosec G114
	// https://github.com/securego/gosec#available-rules

	addr := config.ListenAddress

	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to listen and serve on %s", config.ListenAddress), "error", err)
		os.Exit(1)
	}
}
