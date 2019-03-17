package main

import (
	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/mischief/gopf"
)

var (
	Namespace = "pf"

	fdno = flag.Int("pf.fd", -1, "if set, use this fd for pf ioctls")
)

type PfExporter struct {
	fw pf.Pf

	metrics map[string]*prometheus.Desc
}

// Collect implements the prometheus.Collector interface.
func (e *PfExporter) Collect(ch chan<- prometheus.Metric) {
	stats, err := e.fw.Stats()
	if err != nil {
		log.Errorf("failed to get pf stats: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(e.metrics["state_total"], prometheus.GaugeValue, float64(stats.StateCount()))
	ch <- prometheus.MustNewConstMetric(e.metrics["state_searches"], prometheus.CounterValue, float64(stats.StateSearches()))
	ch <- prometheus.MustNewConstMetric(e.metrics["state_inserts"], prometheus.CounterValue, float64(stats.StateInserts()))
	ch <- prometheus.MustNewConstMetric(e.metrics["state_removals"], prometheus.CounterValue, float64(stats.StateRemovals()))

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
		log.Errorf("failed to get queue stats: %v", err)
		return
	}

	for _, queue := range queues {
		ch <- prometheus.MustNewConstMetric(e.metrics["queue_xmit_packets"], prometheus.CounterValue, float64(queue.Stats.TransmitPackets), queue.Name, queue.IfName)
		ch <- prometheus.MustNewConstMetric(e.metrics["queue_xmit_bytes"], prometheus.CounterValue, float64(queue.Stats.TransmitBytes), queue.Name, queue.IfName)
		ch <- prometheus.MustNewConstMetric(e.metrics["queue_dropped_packets"], prometheus.CounterValue, float64(queue.Stats.DroppedPackets), queue.Name, queue.IfName)
		ch <- prometheus.MustNewConstMetric(e.metrics["queue_dropped_bytes"], prometheus.CounterValue, float64(queue.Stats.DroppedBytes), queue.Name, queue.IfName)
	}
}

func getExporter() (*PfExporter, error) {
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

	// only openbsd has queue support..

	exp := &PfExporter{fw: fw}

	exp.metrics = map[string]*prometheus.Desc{
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
	}

	return exp, nil
}
