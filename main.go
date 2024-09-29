package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mischief/gopf"
)

var (
	Namespace = "pf"
)

type PfExporter struct {
	mu sync.Mutex
	fw pf.Pf

	metrics map[string]*prometheus.Desc
}

// Describe implements the prometheus.Collector interface.
func (e *PfExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.metrics {
		ch <- m
	}
}

// Collect implements the prometheus.Collector interface.
func (e *PfExporter) Collect(ch chan<- prometheus.Metric) {
	e.mu.Lock()
	defer e.mu.Unlock()

	stats, err := e.fw.Stats()
	if err != nil {
		log.Printf("failed to get pf stats: %v", err)
	} else {
		ch <- prometheus.MustNewConstMetric(e.metrics["state_total"], prometheus.GaugeValue, float64(stats.StateCount()))
		ch <- prometheus.MustNewConstMetric(e.metrics["state_searches"], prometheus.CounterValue, float64(stats.StateSearches()))
		ch <- prometheus.MustNewConstMetric(e.metrics["state_inserts"], prometheus.CounterValue, float64(stats.StateInserts()))
		ch <- prometheus.MustNewConstMetric(e.metrics["state_removals"], prometheus.CounterValue, float64(stats.StateRemovals()))
	}

	ifstats := stats.IfStats()
	if ifstats != nil {
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_bytes_in"], prometheus.CounterValue, float64(ifstats.IPv4.BytesIn))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_bytes_out"], prometheus.CounterValue, float64(ifstats.IPv4.BytesOut))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_in_passed"], prometheus.CounterValue, float64(ifstats.IPv4.PacketsInPassed))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_in_blocked"], prometheus.CounterValue, float64(ifstats.IPv4.PacketsInBlocked))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_out_passed"], prometheus.CounterValue, float64(ifstats.IPv4.PacketsOutPassed))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_out_blocked"], prometheus.CounterValue, float64(ifstats.IPv4.PacketsOutBlocked))

		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_bytes_in"], prometheus.CounterValue, float64(ifstats.IPv6.BytesIn))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_bytes_out"], prometheus.CounterValue, float64(ifstats.IPv6.BytesOut))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_in_passed"], prometheus.CounterValue, float64(ifstats.IPv6.PacketsInPassed))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_in_blocked"], prometheus.CounterValue, float64(ifstats.IPv6.PacketsInBlocked))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_out_passed"], prometheus.CounterValue, float64(ifstats.IPv6.PacketsOutPassed))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_out_blocked"], prometheus.CounterValue, float64(ifstats.IPv6.PacketsOutBlocked))
	}

	queues, err := e.fw.Queues()
	if err != nil {
		log.Printf("failed to get queue stats: %v", err)
	} else {
		for _, queue := range queues {
			ch <- prometheus.MustNewConstMetric(e.metrics["queue_xmit_packets"], prometheus.CounterValue, float64(queue.Stats.TransmitPackets), queue.Name, queue.IfName)
			ch <- prometheus.MustNewConstMetric(e.metrics["queue_xmit_bytes"], prometheus.CounterValue, float64(queue.Stats.TransmitBytes), queue.Name, queue.IfName)
			ch <- prometheus.MustNewConstMetric(e.metrics["queue_dropped_packets"], prometheus.CounterValue, float64(queue.Stats.DroppedPackets), queue.Name, queue.IfName)
			ch <- prometheus.MustNewConstMetric(e.metrics["queue_dropped_bytes"], prometheus.CounterValue, float64(queue.Stats.DroppedBytes), queue.Name, queue.IfName)
		}
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
		metrics: map[string]*prometheus.Desc{
			"state_total": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "state", "total"),
				"Number of pf states.",
				nil,
				nil),
			"state_searches": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "state", "searches_total"),
				"Number of pf state searches.",
				nil, nil),
			"state_inserts": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "state", "inserts_total"),
				"Number of pf state inserts.",
				nil,
				nil),
			"state_removals": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "state", "removals_total"),
				"Number of pf state removals.",
				nil,
				nil),

			"ipv4_bytes_in": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv4", "bytes_in_total"),
				"Number of bytes in on the pf loginterface over IPv4.",
				nil,
				nil),
			"ipv4_bytes_out": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv4", "bytes_out_total"),
				"Number of bytes out on the pf loginterface over IPv4.",
				nil,
				nil),
			"ipv4_packets_in_passed": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv4", "packets_in_passed_total"),
				"Number of packets passed in on the pf loginterface over IPv4.",
				nil,
				nil),
			"ipv4_packets_in_blocked": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv4", "packets_in_blocked_total"),
				"Number of packets blocked in on the pf loginterface over IPv4.",
				nil,
				nil),
			"ipv4_packets_out_passed": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv4", "packets_out_passed_total"),
				"Number of packets passed out on the pf loginterface over IPv4.",
				nil,
				nil),
			"ipv4_packets_out_blocked": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv4", "packets_out_blocked_total"),
				"Number of packets blocked out on the pf loginterface over IPv4.",
				nil,
				nil),

			"ipv6_bytes_in": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv6", "bytes_in_total"),
				"Number of bytes in on the pf loginterface over IPv6.",
				nil,
				nil),
			"ipv6_bytes_out": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv6", "bytes_out_total"),
				"Number of bytes out on the pf loginterface over IPv6.",
				nil,
				nil),
			"ipv6_packets_in_passed": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv6", "packets_in_passed_total"),
				"Number of packets passed in on the pf loginterface over IPv6.",
				nil,
				nil),
			"ipv6_packets_in_blocked": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv6", "packets_in_blocked_total"),
				"Number of packets blocked in on the pf loginterface over IPv6.",
				nil,
				nil),
			"ipv6_packets_out_passed": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv6", "packets_out_passed_total"),
				"Number of packets passed out on the pf loginterface over IPv6.",
				nil,
				nil),
			"ipv6_packets_out_blocked": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "ipv6", "packets_out_blocked_total"),
				"Number of packets blocked out on the pf loginterface over IPv6.",
				nil,
				nil),

			"queue_xmit_packets": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "stats", "queue_transmitted_packets_total"),
				"Number of transmitted packets in a queue partitioned by queue name and interface.",
				[]string{"queue", "interface"},
				nil),
			"queue_xmit_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "stats", "queue_transmitted_bytes_total"),
				"Number of transmitted bytes in a queue partitioned by queue name and interface.",
				[]string{"queue", "interface"},
				nil),
			"queue_dropped_packets": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "stats", "queue_dropped_packets_total"),
				"Number of dropped packets in a queue partitioned by queue name and interface.",
				[]string{"queue", "interface"},
				nil),
			"queue_dropped_bytes": prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, "stats", "queue_dropped_bytes_total"),
				"Number of dropped bytes in a queue partitioned by queue name and interface.",
				[]string{"queue", "interface"},
				nil),
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

	reg := prometheus.NewRegistry()
	reg.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listenAddress)

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorLog: log.Default(), Registry: reg}))
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
