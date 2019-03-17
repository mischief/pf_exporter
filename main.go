package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	Namespace = "pf"
)

// Describe implements the prometheus.Collector interface.
func (e *PfExporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(e, ch)
}

func NewPfExporter() (*PfExporter, error) {
	exp, err := getExporter()
	if err != nil {
		return nil, err
	}

	for k, v := range map[string]*prometheus.Desc{
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
	} {
		exp.metrics[k] = v
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
