package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"radius-exporter/collector"
	"radius-exporter/config"

	log "github.com/sirupsen/logrus"

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
		log.Error("No target specified.")
		return
	}

	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		http.Error(w, "No module specified", http.StatusBadRequest)
		log.Debug("No module specified.")
		return
	}
	module, ok := exporterConfig.Modules[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown module %q", moduleName), http.StatusBadRequest)
		log.Debugf("Unknown module %q", moduleName)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewCollector(&target, &module))
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, r)
}

func configureLog() {
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)

	// Add timestamp
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
}

func init() {
	flag.Parse()
}

func main() {
	flag.Parse()

	if *showVer {
		fmt.Printf("Version: %s\n", exporterVersion)
		fmt.Printf("SHA: %s\n", exporterSha)
		os.Exit(0)
	}
	log.Info("Starting RADIUS exporter")

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
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
