package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/go-freebsd/pf"
)

type PfExporter struct {
	fw *pf.Handle

	metrics map[string]*prometheus.Desc
}

// Collect implements the prometheus.Collector interface.
func (e *PfExporter) Collect(ch chan<- prometheus.Metric) {
	var stats pf.Statistics

	if err := e.fw.UpdateStatistics(&stats); err != nil {
		log.Errorf("failed to get pf stats: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(e.metrics["state_total"], prometheus.GaugeValue, float64(stats.CounterStates()))
	ch <- prometheus.MustNewConstMetric(e.metrics["state_searches"], prometheus.CounterValue, float64(stats.CounterStateSearch()))
	ch <- prometheus.MustNewConstMetric(e.metrics["state_inserts"], prometheus.CounterValue, float64(stats.CounterStateInsert()))
	ch <- prometheus.MustNewConstMetric(e.metrics["state_removals"], prometheus.CounterValue, float64(stats.CounterStateRemovals()))

	if stats.Interface() != "" {
		bstats := stats.Bytes()
		pdrop := stats.PacketsDrop()
		ppass := stats.PacketsPass()

		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_bytes_in"], prometheus.CounterValue, float64(bstats.ReceivedIPv4))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_bytes_out"], prometheus.CounterValue, float64(bstats.SendIPv4))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_in_passed"], prometheus.CounterValue, float64(ppass.ReceivedIPv4))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_in_blocked"], prometheus.CounterValue, float64(pdrop.ReceivedIPv4))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_out_passed"], prometheus.CounterValue, float64(ppass.SendIPv4))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv4_packets_out_blocked"], prometheus.CounterValue, float64(pdrop.SendIPv4))

		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_bytes_in"], prometheus.CounterValue, float64(bstats.ReceivedIPv6))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_bytes_out"], prometheus.CounterValue, float64(bstats.SendIPv6))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_in_passed"], prometheus.CounterValue, float64(ppass.ReceivedIPv6))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_in_blocked"], prometheus.CounterValue, float64(pdrop.ReceivedIPv6))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_out_passed"], prometheus.CounterValue, float64(ppass.SendIPv6))
		ch <- prometheus.MustNewConstMetric(e.metrics["ipv6_packets_out_blocked"], prometheus.CounterValue, float64(pdrop.SendIPv6))
	}
}

func getExporter() (*PfExporter, error) {
	var fw *pf.Handle
	var err error

	fw, err = pf.Open()
	if err != nil {
		return nil, err
	}

	exp := &PfExporter{
		fw:      fw,
		metrics: map[string]*prometheus.Desc{},
	}

	return exp, nil
}
