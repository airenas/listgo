package metrics

import "github.com/prometheus/client_golang/prometheus"

//Register tries to register or reregister metric to prometheus default registry
func Register(m prometheus.Collector) error {
	err := prometheus.Register(m)
	if err != nil {
		prometheus.Unregister(m)
		err = prometheus.Register(m)
	}
	return err
}
