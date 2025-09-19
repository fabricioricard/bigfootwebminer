// +build !monitoring

package monitoring

import (
	"google.golang.org/grpc"

	"github.com/bigchain/bigchaind/btcutil/er"
	"github.com/bigchain/bigchaind/lnd/lncfg"
)

// GetPromInterceptors returns the set of interceptors for Prometheus
// monitoring if monitoring is enabled, else empty slices. Monitoring is
// currently disabled.
func GetPromInterceptors() ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	return []grpc.UnaryServerInterceptor{}, []grpc.StreamServerInterceptor{}
}

// ExportPrometheusMetrics is required for lnd to compile so that Prometheus
// metric exporting can be hidden behind a build tag.
func ExportPrometheusMetrics(_ *grpc.Server, _ lncfg.Prometheus) er.R {
	return er.Errorf("lnd must be built with the monitoring tag to " +
		"enable exporting Prometheus metrics")
}
