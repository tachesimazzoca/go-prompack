package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tachesimazzoca/go-prompack/core"
	"github.com/tachesimazzoca/go-prompack/exporter"

	_ "github.com/go-sql-driver/mysql"
)

func metricHandler(registry *prometheus.Registry) http.Handler {
	return promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
}

func must(b bool, v ...interface{}) {
	if !b {
		log.Fatal(v)
	}
}

func main() {
	var err error

	flgs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfgPath := flgs.String("c", "config.yml", "config")
	err = flgs.Parse(os.Args[1:])
	must(err == nil, err)

	yml, err := ioutil.ReadFile(*cfgPath)
	must(err == nil, err)

	cfg, err := core.NewExporterConfigFromYaml(yml)
	must(err == nil, err)

	ep, err := exporter.NewExporter(cfg)
	must(err == nil, err)

	ctx := context.Background()
	registry := prometheus.NewRegistry()

	err = ep.Start(ctx, registry)
	must(err == nil, err)

	handler := http.NewServeMux()
	handler.Handle("/metrics", metricHandler(registry))
	srv := http.Server{
		Addr:    cfg.Server.Addr,
		Handler: handler,
	}

	log.Printf("Waiting for SIGTERM to stop server")
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	cancel := make(chan struct{})
	go func() {
		for {
			select {
			case <-term:
				log.Println("Received SIGTERM, exiting gracefully...")
				if err := ep.Stop(); err != nil {
					log.Printf("(%p).Stop: %v", ep, err)
				}
				log.Printf("Stopping server: %s", srv.Addr)
				if err := srv.Shutdown(ctx); err != nil {
					log.Printf("srv.Shutdown returned error: %v", err)
				} else {
					log.Printf("srv.Shutdown successfully")
				}
				close(cancel)
				return
			}
		}
	}()

	log.Printf("Server started: %s", srv.Addr)
	err = srv.ListenAndServe()
	must(err == http.ErrServerClosed, err)

	<-cancel
	log.Printf("Server stopped: %s", srv.Addr)
}
