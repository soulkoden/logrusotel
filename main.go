package logrusotel

import (
	"context"
	"fmt"
	"net"
	"runtime"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteljaeger "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

func NewTracerProvider(url, name string, debug bool) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	if !debug {
		host, port, err := net.SplitHostPort(url)
		if err != nil {
			return nil, fmt.Errorf("failed to split host:port: %w", err)
		}

		exporter, err = oteljaeger.New(oteljaeger.WithAgentEndpoint(
			oteljaeger.WithAgentHost(host),
			oteljaeger.WithAgentPort(port)))
	} else {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create new jaeger exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			attribute.String("host.arch", runtime.GOARCH),
			attribute.String("service.name", name),
		))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

// ---

type JaegerHook struct{}

func NewJaegerHook() *JaegerHook {
	return &JaegerHook{}
}

func (t *JaegerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t *JaegerHook) Fire(e *logrus.Entry) error {
	span := trace.SpanFromContext(e.Context)
	if span == nil || !span.IsRecording() {
		return nil
	}

	spanAttributes := make([]attribute.KeyValue, 0, len(e.Data))
	for k, v := range e.Data {
		switch v := v.(type) {
		case bool:
			spanAttributes = append(spanAttributes, attribute.Bool(k, v))
		case []bool:
			spanAttributes = append(spanAttributes, attribute.BoolSlice(k, v))
		case int:
			spanAttributes = append(spanAttributes, attribute.Int(k, v))
		case []int:
			spanAttributes = append(spanAttributes, attribute.IntSlice(k, v))
		case int64:
			spanAttributes = append(spanAttributes, attribute.Int64(k, v))
		case []int64:
			spanAttributes = append(spanAttributes, attribute.Int64Slice(k, v))
		case float64:
			spanAttributes = append(spanAttributes, attribute.Float64(k, v))
		case []float64:
			spanAttributes = append(spanAttributes, attribute.Float64Slice(k, v))
		case string:
			spanAttributes = append(spanAttributes, attribute.String(k, v))
		case []string:
			spanAttributes = append(spanAttributes, attribute.StringSlice(k, v))
		case fmt.Stringer:
			spanAttributes = append(spanAttributes, attribute.String(k, v.String()))
		case error:
			span.SetStatus(codes.Error, v.Error())
			span.RecordError(v)
		default:
			spanAttributes = append(spanAttributes, attribute.String(k, spew.Sdump(v)))
		}
	}

	span.AddEvent(e.Message, trace.WithAttributes(spanAttributes...))

	return nil
}
