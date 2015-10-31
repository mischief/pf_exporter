package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/mischief/gopf"
)

var (
	Namespace = "pf"
)

type PfExporter struct {
	fw       pf.Pf
	gauges   map[string]prometheus.Gauge
	counters map[string]prometheus.Counter
}

// Describe implements the prometheus.Collector interface.
func (e *PfExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.gauges {
		m.Describe(ch)
	}

	for _, m := range e.counters {
		m.Describe(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (e *PfExporter) Collect(ch chan<- prometheus.Metric) {
	stats, err := e.fw.Stats()
	if err != nil {
		log.Errorf("failed to get pf stats: %v", err)
		return
	}

	e.gauges["state_total"].Set(float64(stats.StateCount()))
	e.counters["state_searches"].Set(float64(stats.StateSearches()))
	e.counters["state_inserts"].Set(float64(stats.StateInserts()))
	e.counters["state_removals"].Set(float64(stats.StateRemovals()))

	for _, m := range e.gauges {
		m.Collect(ch)
	}

	for _, m := range e.counters {
		m.Collect(ch)
	}
}

func NewPfExporter() (*PfExporter, error) {
	fw, err := pf.Open()
	if err != nil {
		return nil, err
	}

	exp := &PfExporter{
		fw: fw,
		gauges: map[string]prometheus.Gauge{
			"state_total": prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: Namespace,
				Subsystem: "state",
				Name:      "total",
				Help:      "Number of pf states.",
			}),
		},
		counters: map[string]prometheus.Counter{
			"state_searches": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "state",
				Name:      "searches_total",
				Help:      "Number of pf state searches.",
			}),
			"state_inserts": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "state",
				Name:      "inserts_total",
				Help:      "Number of pf state inserts.",
			}),
			"state_removals": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "state",
				Name:      "removals_total",
				Help:      "Number of pf state removals.",
			}),
		},
	}

	return exp, nil
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9107", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	exporter, err := NewPfExporter()
	if err != nil {
		log.Fatalf("Failed to create pf exporter: %v", err)
	}

	prometheus.MustRegister(exporter)

	log.Infof("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>pf Exporter</title></head>
             <body>
             <h1>pf Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
