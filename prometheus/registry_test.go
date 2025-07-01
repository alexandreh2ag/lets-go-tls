package prometheus

import (
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	registryProm := prometheus.NewRegistry()
	registry := NewRegistry("foo", registryProm)
	assert.NotNil(t, registry)
}

func Test_stdRegistry_FormatName(t *testing.T) {
	name := "foo"
	tests := []struct {
		name      string
		namespace string
		want      string
	}{
		{
			name:      "Success",
			namespace: "foo",
			want:      "foo_foo",
		},
		{
			name:      "SuccessNamespaceEmpty",
			namespace: "",
			want:      "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := &stdRegistry{
				namespace: tt.namespace,
			}
			assert.Equalf(t, tt.want, sr.FormatName(name), "FormatName(%v)", name)
		})
	}
}

func Test_stdRegistry_MustAddGauge(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:     registryProm,
		namespace:    "",
		gaugeMetrics: map[string]prometheus.Gauge{},
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "Foo",
	})

	sr.MustAddGauge(name, metric)
	assert.Len(t, sr.gaugeMetrics, 1)
}

func Test_stdRegistry_MustAddGaugeCertificate_NotExist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:            registryProm,
		namespace:           "",
		certificatesMetrics: map[string]prometheus.Gauge{},
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "Foo",
	})

	sr.MustAddGaugeCertificate(name, metric)
	assert.Len(t, sr.certificatesMetrics, 1)
}

func Test_stdRegistry_MustAddGaugeCertificate_Exist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:            registryProm,
		namespace:           "",
		certificatesMetrics: map[string]prometheus.Gauge{},
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "Foo",
	})

	sr.MustAddGaugeCertificate(name, metric)
	sr.MustAddGaugeCertificate(name, metric)
	assert.Len(t, sr.certificatesMetrics, 1)
}

func Test_stdRegistry_MustAddCounter(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:       registryProm,
		namespace:      "",
		counterMetrics: map[string]prometheus.Counter{},
	}

	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo",
		Help: "Foo",
	})

	sr.MustAddCounter(name, metric)
	assert.Len(t, sr.counterMetrics, 1)
}

func Test_stdRegistry_MustDeleteGauge(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:  registryProm,
		namespace: "",
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "Foo",
	})
	sr.gaugeMetrics = map[string]prometheus.Gauge{name: metric}
	assert.Len(t, sr.gaugeMetrics, 1)
	sr.MustDeleteGauge(name)
	assert.Len(t, sr.gaugeMetrics, 0)
}

func Test_stdRegistry_MustDeleteGaugeCertificate(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:  registryProm,
		namespace: "",
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "Foo",
	})
	sr.certificatesMetrics = map[string]prometheus.Gauge{name: metric}
	assert.Len(t, sr.certificatesMetrics, 1)
	sr.MustDeleteGaugeCertificate(name)
	assert.Len(t, sr.certificatesMetrics, 0)
}

func Test_stdRegistry_MustDeleteCounter(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:  registryProm,
		namespace: "",
	}

	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo",
		Help: "Foo",
	})
	sr.counterMetrics = map[string]prometheus.Counter{name: metric}
	assert.Len(t, sr.counterMetrics, 1)
	sr.MustDeleteCounter(name)
	assert.Len(t, sr.counterMetrics, 0)
}

func Test_stdRegistry_MustGetGauge_SuccessExist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:  registryProm,
		namespace: "",
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "Foo",
	})
	sr.gaugeMetrics = map[string]prometheus.Gauge{name: metric}
	got := sr.MustGetGauge(name)
	assert.Equal(t, metric, got)
}

func Test_stdRegistry_MustGetGauge_SuccessNotExist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:     registryProm,
		namespace:    "",
		gaugeMetrics: map[string]prometheus.Gauge{},
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
	})
	got := sr.MustGetGauge(name)
	assert.Equal(t, metric, got)
}

func Test_stdRegistry_MustGetGaugeCertificate_SuccessExist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:  registryProm,
		namespace: "",
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
		Help: "Foo",
	})
	sr.certificatesMetrics = map[string]prometheus.Gauge{name: metric}
	got := sr.MustGetGaugeCertificate(name)
	assert.Equal(t, metric, got)
}

func Test_stdRegistry_MustGetGaugeCertificate_SuccessNotExist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:            registryProm,
		namespace:           "",
		certificatesMetrics: map[string]prometheus.Gauge{},
	}

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
	})
	got := sr.MustGetGaugeCertificate(name)
	assert.Equal(t, metric, got)
}

func Test_stdRegistry_CreateGaugeCertificate(t *testing.T) {
	cert := &types.Certificate{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com", "a.example.com"}}
	want := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo_tls_certs_not_after",
		Help: "Certificate expiration timestamp",
		ConstLabels: map[string]string{
			"identifier": cert.Identifier,
			"cn":         cert.Main,
			"sans":       strings.Join([]string{"a.example.com", "example.com"}, ","),
		},
	})
	sr := &stdRegistry{
		namespace: "foo",
	}
	assert.Equalf(t, want, sr.CreateGaugeCertificate(cert), "CreateGaugeCertificate(%v)", cert)
}

func Test_stdRegistry_MustGetCounter_SuccessExist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:  registryProm,
		namespace: "",
	}

	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo",
		Help: "Foo",
	})
	sr.counterMetrics = map[string]prometheus.Counter{name: metric}
	got := sr.MustGetCounter(name)
	assert.Equal(t, metric, got)
}

func Test_stdRegistry_MustGetCounter_SuccessNotExist(t *testing.T) {
	name := "foo"
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		Registry:       registryProm,
		namespace:      "",
		counterMetrics: map[string]prometheus.Counter{},
	}

	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo",
	})
	got := sr.MustGetCounter(name)
	assert.Equal(t, metric.Desc().String(), got.Desc().String())
}

func Test_stdRegistry_GetGaugeCertificates(t *testing.T) {
	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "foo",
	})
	sr := &stdRegistry{
		namespace:           "",
		certificatesMetrics: map[string]prometheus.Gauge{"foo": metric},
	}

	got := sr.GetGaugeCertificates()
	assert.Equal(t, sr.certificatesMetrics, got)
}

func Test_stdRegistry_RegisterNewCertificateMetrics(t *testing.T) {
	registryProm := prometheus.NewRegistry()
	sr := &stdRegistry{
		namespace:           "",
		Registry:            registryProm,
		certificatesMetrics: map[string]prometheus.Gauge{},
	}
	cert := &types.Certificate{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}}
	sr.RegisterNewCertificateMetrics(cert)
	assert.Len(t, sr.certificatesMetrics, 1)
}

func Test_stdRegistry_UpdateCertificatesMetrics(t *testing.T) {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tls_certs_not_after",
		Help: "Certificate expiration timestamp",
		ConstLabels: map[string]string{
			"identifier": "foo",
			"cn":         "example.com",
			"sans":       "example.com",
		},
	})
	certDate := time.Date(2035, time.April, 7, 00, 41, 15, 0, time.UTC)
	tests := []struct {
		name                string
		certificates        types.Certificates
		certificatesMetrics map[string]prometheus.Gauge
		checkFn             func(t *testing.T, sr *stdRegistry)
	}{
		{
			name: "SuccessEmptyCertificates",
			checkFn: func(t *testing.T, sr *stdRegistry) {
				assert.Len(t, sr.certificatesMetrics, 0)
			},
		},
		{
			name:                "SuccessUpdateCertificate",
			certificatesMetrics: map[string]prometheus.Gauge{"foo": gauge},
			certificates:        types.Certificates{{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}, ExpirationDate: certDate}},
			checkFn: func(t *testing.T, sr *stdRegistry) {
				assert.Len(t, sr.certificatesMetrics, 1)
				value := testutil.ToFloat64(sr.certificatesMetrics["foo"])
				assert.Equal(t, float64(certDate.Unix()), value)
			},
		},
		{
			name:                "SuccessRemoveCertificate",
			certificatesMetrics: map[string]prometheus.Gauge{"foo": gauge},
			certificates:        types.Certificates{},
			checkFn: func(t *testing.T, sr *stdRegistry) {
				assert.Len(t, sr.certificatesMetrics, 0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registryProm := prometheus.NewRegistry()
			sr := &stdRegistry{
				namespace:           "",
				Registry:            registryProm,
				certificatesMetrics: tt.certificatesMetrics,
			}
			sr.UpdateCertificatesMetrics(tt.certificates)
			tt.checkFn(t, sr)
		})
	}
}
