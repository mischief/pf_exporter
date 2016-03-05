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
	fw          pf.Pf
	gauges      map[string]prometheus.Gauge
	counters    map[string]prometheus.Counter
	countervecs map[string]*prometheus.CounterVec
}

// Describe implements the prometheus.Collector interface.
func (e *PfExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.gauges {
		m.Describe(ch)
	}

	for _, m := range e.counters {
		m.Describe(ch)
	}

	for _, m := range e.countervecs {
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

	ifstats := stats.IfStats()
	if ifstats != nil {
		e.counters["ipv4_bytes_in"].Set(float64(ifstats.IPv4.BytesIn))
		e.counters["ipv4_bytes_out"].Set(float64(ifstats.IPv4.BytesOut))
		e.counters["ipv4_packets_in_passed"].Set(float64(ifstats.IPv4.PacketsInPassed))
		e.counters["ipv4_packets_in_blocked"].Set(float64(ifstats.IPv4.PacketsInBlocked))
		e.counters["ipv4_packets_out_passed"].Set(float64(ifstats.IPv4.PacketsOutPassed))
		e.counters["ipv4_packets_out_blocked"].Set(float64(ifstats.IPv4.PacketsOutBlocked))

		e.counters["ipv6_bytes_in"].Set(float64(ifstats.IPv6.BytesIn))
		e.counters["ipv6_bytes_out"].Set(float64(ifstats.IPv6.BytesOut))
		e.counters["ipv6_packets_in_passed"].Set(float64(ifstats.IPv6.PacketsInPassed))
		e.counters["ipv6_packets_in_blocked"].Set(float64(ifstats.IPv6.PacketsInBlocked))
		e.counters["ipv6_packets_out_passed"].Set(float64(ifstats.IPv6.PacketsOutPassed))
		e.counters["ipv6_packets_out_blocked"].Set(float64(ifstats.IPv6.PacketsOutBlocked))
	}

	queues, err := e.fw.Queues()
	if err != nil {
		log.Errorf("failed to get queue stats: %v", err)
		return
	}

	for _, queue := range queues {
		e.countervecs["queue_xmit_packets"].WithLabelValues(queue.Name, queue.IfName).Set(float64(queue.Stats.TransmitPackets))
		e.countervecs["queue_xmit_bytes"].WithLabelValues(queue.Name, queue.IfName).Set(float64(queue.Stats.TransmitBytes))
		e.countervecs["queue_dropped_packets"].WithLabelValues(queue.Name, queue.IfName).Set(float64(queue.Stats.DroppedPackets))
		e.countervecs["queue_dropped_bytes"].WithLabelValues(queue.Name, queue.IfName).Set(float64(queue.Stats.DroppedBytes))
	}

	for _, m := range e.gauges {
		m.Collect(ch)
	}

	for _, m := range e.counters {
		m.Collect(ch)
	}

	for _, m := range e.countervecs {
		m.Collect(ch)
	}
}

func NewPfExporter() (*PfExporter, error) {
	var fw pf.Pf
	var err error

	if *fdno != -1 {
		fw = pf.OpenFD(uintptr(*fdno))
	} else {
		fw, err = pf.Open()
		if err != nil {
			return nil, err
		}
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

			"ipv4_bytes_in": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv4_bytes_in_total",
				Help:      "Number of bytes in on the pf loginterface over IPv4.",
			}),
			"ipv4_bytes_out": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv4_bytes_out_total",
				Help:      "Number of bytes out on the pf loginterface over IPv4.",
			}),
			"ipv4_packets_in_passed": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv4_packets_in_passed_total",
				Help:      "Number of packets passed in on the pf loginterface over IPv4.",
			}),
			"ipv4_packets_in_blocked": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv4_packets_in_blocked_total",
				Help:      "Number of packets blocked in on the pf loginterface over IPv4.",
			}),
			"ipv4_packets_out_passed": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv4_packets_out_passed_total",
				Help:      "Number of packets passed out on the pf loginterface over IPv4.",
			}),
			"ipv4_packets_out_blocked": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv4_packets_out_blocked_total",
				Help:      "Number of packets blocked out on the pf loginterface over IPv4.",
			}),

			"ipv6_bytes_in": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv6_bytes_in_total",
				Help:      "Number of bytes in on the pf loginterface over IPv6.",
			}),
			"ipv6_bytes_out": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv6_bytes_out_total",
				Help:      "Number of bytes out on the pf loginterface over IPv6.",
			}),
			"ipv6_packets_in_passed": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv6_packets_in_passed_total",
				Help:      "Number of packets passed in on the pf loginterface over IPv6.",
			}),
			"ipv6_packets_in_blocked": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv6_packets_in_blocked_total",
				Help:      "Number of packets blocked in on the pf loginterface over IPv6.",
			}),
			"ipv6_packets_out_passed": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv6_packets_out_passed_total",
				Help:      "Number of packets passed out on the pf loginterface over IPv6.",
			}),
			"ipv6_packets_out_blocked": prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "ipv6_packets_out_blocked_total",
				Help:      "Number of packets blocked out on the pf loginterface over IPv6.",
			}),
		},

		countervecs: map[string]*prometheus.CounterVec{
			"queue_xmit_packets": prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "queue_transmitted_packets_total",
				Help:      "Number of transmitted packets in a queue partitioned by queue name and interface",
			},
				[]string{"queue", "interface"},
			),
			"queue_xmit_bytes": prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "queue_transmitted_bytes_total",
				Help:      "Number of transmitted bytes in a queue partitioned by queue name and interface",
			},
				[]string{"queue", "interface"},
			),
			"queue_dropped_packets": prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "queue_dropped_packets_total",
				Help:      "Number of dropped packets in a queue partitioned by queue name and interface",
			},
				[]string{"queue", "interface"},
			),
			"queue_dropped_bytes": prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "stats",
				Name:      "queue_dropped_bytes_total",
				Help:      "Number of dropped bytes in a queue partitioned by queue name and interface",
			},
				[]string{"queue", "interface"},
			),
		},
	}

	return exp, nil
}

var (
	listenAddress = flag.String("web.listen-address", ":9107", "Address to listen on for web interface and telemetry.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	fdno          = flag.Int("pf.fd", -1, "if set, use this fd for pf ioctls")
)

func main() {
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
