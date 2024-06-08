package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/devon-mar/radius-exporter/collector"
	"github.com/devon-mar/radius-exporter/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	showVer     = flag.Bool("version", false, "Show the version")
	configPath  = flag.String("config", "config.yml", "Path to the config file.")
	listenAddr  = flag.String("web.listen-address", ":9881", "Address on which to expose metrics and web interface.")
	metricsPath = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	logLevel    = flag.String("log.level", "info", "The log level.")

	exporterConfig *config.Config

	exporterVersion = "DEVELOPMENT"
	exporterSha     = "abcd"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "No target specified", http.StatusBadRequest)
		slog.Error("No target specified.")
		return
	}

	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		http.Error(w, "No module specified", http.StatusBadRequest)
		slog.Debug("No module specified.")
		return
	}
	module, ok := exporterConfig.Modules[moduleName]
	if !ok {
		http.Error(w, "Unknown module", http.StatusBadRequest)
		slog.Debug("Unknown module", "module", moduleName)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewCollector(&target, &module))
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, r)
}

func configureLog() {
	var slogLevel slog.Level

	if err := slogLevel.UnmarshalText([]byte(*logLevel)); err != nil {
		slog.Error("error parsing log level", "err", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel}))
	slog.SetDefault(logger)
}

func main() {
	flag.Parse()

	if *showVer {
		fmt.Printf("Version: %s\n", exporterVersion)
		fmt.Printf("SHA: %s\n", exporterSha)
		os.Exit(0)
	}
	slog.Info("Starting RADIUS exporter")

	configureLog()

	c, err := config.LoadFromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	exporterConfig = c

	http.Handle(*metricsPath, http.HandlerFunc(metricsHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
				<head><title>RADIUS Exporter</title></head>
				<body>
					<h1>RADIUS Exporter</h1>
					<a href="` + *metricsPath + `">Metrics</a></p>
				</body>
				</html>`))
	})

	server := http.Server{Addr: *listenAddr}
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)

		signal.Notify(sigCh, os.Interrupt)
		sig := <-sigCh
		slog.Warn("received signal", "signal", sig)

		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("HTTP server shutdown error", "err", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("HTTP server ListenAndServe", "err", err)
	}

	<-idleConnsClosed
}
