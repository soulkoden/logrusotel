package pkg

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func NewTracer(serviceName string, queueSize int) (opentracing.Tracer, io.Closer, error) {
	cfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: os.Getenv("JAEGER_AGENT_HOST_PORT"),
			QueueSize:          queueSize,
		},
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.NullLogger))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new tracer: %w", err)
	}

	return tracer, closer, nil
}

func MustTracerCloser(serviceName string, queueSize int) (opentracing.Tracer, io.Closer) {
	tracer, closer, err := NewTracer(serviceName, queueSize)
	if err != nil {
		slog.With("error", err).Error("failed to create tracer")
	}
	return tracer, closer
}

func MustClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		slog.With("error", err).Error("failed to close tracer")
	}
}
