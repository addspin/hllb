package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Общее количество запросов по типу (A, NS, ...)
	QueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hllb_queries_total",
			Help: "Total DNS queries by type",
		},
		[]string{"type"},
	)

	// Ответы по rcode (NOERROR, NXDOMAIN, SERVFAIL)
	ResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hllb_responses_total",
			Help: "Total DNS responses by rcode",
		},
		[]string{"rcode"},
	)

	// Запросы ушедшие на forward
	ForwardTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "hllb_forward_total",
			Help: "Total forwarded DNS queries",
		},
	)

	// Ошибки форвардинга
	ForwardErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "hllb_forward_errors_total",
			Help: "Total forwarding errors",
		},
	)

	// Латенция обработки запросов
	QueryDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "hllb_query_duration_seconds",
			Help:    "DNS query processing duration",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
		},
	)

	// Размер пула живых хостов
	PoolSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hllb_pool_size",
			Help: "Number of alive hosts in check pool",
		},
	)
)

func Init() {
	prometheus.MustRegister(
		QueriesTotal,
		ResponsesTotal,
		ForwardTotal,
		ForwardErrorsTotal,
		QueryDuration,
		PoolSize,
	)
}

func ServeHTTP(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(addr, nil)
}
