package prometheus

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

type Registry interface {
	prometheus.Registerer
	prometheus.Gatherer

	FormatName(name string) string

	MustAddGauge(name string, metric prometheus.Gauge)
	MustDeleteGauge(name string)
	MustGetGauge(name string) prometheus.Gauge

	MustAddGaugeCertificate(name string, metric prometheus.Gauge)
	MustDeleteGaugeCertificate(name string)
	MustGetGaugeCertificate(name string) prometheus.Gauge
	GetGaugeCertificates() map[string]prometheus.Gauge
	CreateGaugeCertificate(certificate *types.Certificate) prometheus.Gauge

	MustAddCounter(name string, metric prometheus.Counter)
	MustDeleteCounter(name string)
	MustGetCounter(name string) prometheus.Counter
}

type stdRegistry struct {
	*prometheus.Registry
	namespace string

	gaugeMetrics        map[string]prometheus.Gauge
	certificatesMetrics map[string]prometheus.Gauge
	counterMetrics      map[string]prometheus.Counter
}

func NewRegistry(namespace string, registryProm *prometheus.Registry) Registry {
	return &stdRegistry{
		Registry:            registryProm,
		namespace:           namespace,
		gaugeMetrics:        make(map[string]prometheus.Gauge),
		certificatesMetrics: make(map[string]prometheus.Gauge),
		counterMetrics:      make(map[string]prometheus.Counter),
	}
}

func (sr *stdRegistry) FormatName(name string) string {
	if sr.namespace != "" {
		return fmt.Sprintf("%s_%s", sr.namespace, name)
	}
	return name
}

func (sr *stdRegistry) MustAddGauge(name string, metric prometheus.Gauge) {
	sr.gaugeMetrics[name] = metric
	sr.MustRegister(metric)
}

func (sr *stdRegistry) MustDeleteGauge(name string) {
	if metric, ok := sr.gaugeMetrics[name]; ok {
		sr.Unregister(metric)
		delete(sr.gaugeMetrics, name)
	}
}

func (sr *stdRegistry) MustGetGauge(name string) prometheus.Gauge {
	if metric, ok := sr.gaugeMetrics[name]; ok {
		return metric
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: sr.FormatName(name),
	})
	sr.MustAddGauge(name, metric)
	return metric
}

func (sr *stdRegistry) MustAddGaugeCertificate(name string, metric prometheus.Gauge) {
	if _, ok := sr.certificatesMetrics[name]; !ok {
		sr.certificatesMetrics[name] = metric
		sr.MustRegister(metric)
	}
}

func (sr *stdRegistry) MustDeleteGaugeCertificate(name string) {
	if metric, ok := sr.certificatesMetrics[name]; ok {
		sr.Unregister(metric)
		delete(sr.certificatesMetrics, name)
	}
}

func (sr *stdRegistry) MustGetGaugeCertificate(name string) prometheus.Gauge {
	if metric, ok := sr.certificatesMetrics[name]; ok {
		return metric
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: sr.FormatName(name),
	})
	sr.MustAddGaugeCertificate(name, metric)
	return metric
}

func (sr *stdRegistry) CreateGaugeCertificate(certificate *types.Certificate) prometheus.Gauge {
	certificate.Domains.Sort()
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Name: sr.FormatName("tls_certs_not_after"),
		Help: "Certificate expiration timestamp",
		ConstLabels: map[string]string{
			"identifier": certificate.Identifier,
			"cn":         certificate.Main,
			"sans":       strings.Join(certificate.Domains.ToStringSlice(), ","),
		},
	})
}

func (sr *stdRegistry) GetGaugeCertificates() map[string]prometheus.Gauge {
	return sr.certificatesMetrics
}

func (sr *stdRegistry) MustAddCounter(name string, metric prometheus.Counter) {
	sr.counterMetrics[name] = metric
	sr.MustRegister(metric)
}

func (sr *stdRegistry) MustDeleteCounter(name string) {
	if metric, ok := sr.counterMetrics[name]; ok {
		sr.Unregister(metric)
		delete(sr.counterMetrics, name)
	}
}

func (sr *stdRegistry) MustGetCounter(name string) prometheus.Counter {
	if metric, ok := sr.counterMetrics[name]; ok {
		return metric
	}

	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: sr.FormatName(name),
	})
	sr.MustAddCounter(name, metric)
	return metric
}
