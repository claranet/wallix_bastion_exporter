package main

import (
	"log"
	"net/http"

	"github.com/claranet/wallix_bastion_exporter/config"
	"github.com/claranet/wallix_bastion_exporter/exporter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// Test
	// fmt.Println(viper.GetString("listen-address"))
	// fmt.Println(config.WallixUsername)
	// fmt.Println(config.SkipVerify)

	wallixExporter := exporter.NewExporter(cfg)
	prometheus.MustRegister(wallixExporter)

	http.Handle(cfg.TelemetryPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, cfg.TelemetryPath, http.StatusPermanentRedirect)
	})
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, nil))
}
