package logrusotel

import (
	"context"
	"fmt"
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
		exporter, err = oteljaeger.New(oteljaeger.WithCollectorEndpoint(oteljaeger.WithEndpoint(url)))
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
	if span == nil {
		return nil
	}

	if !span.IsRecording() {
		return nil
	}

	span.AddEvent(e.Message)

	for k, v := range e.Data {
		switch v := v.(type) {
		case bool:
			span.SetAttributes(attribute.Bool(k, v))
		case []bool:
			span.SetAttributes(attribute.BoolSlice(k, v))
		case int:
			span.SetAttributes(attribute.Int(k, v))
		case []int:
			span.SetAttributes(attribute.IntSlice(k, v))
		case int64:
			span.SetAttributes(attribute.Int64(k, v))
		case []int64:
			span.SetAttributes(attribute.Int64Slice(k, v))
		case float64:
			span.SetAttributes(attribute.Float64(k, v))
		case []float64:
			span.SetAttributes(attribute.Float64Slice(k, v))
		case string:
			span.SetAttributes(attribute.String(k, v))
		case []string:
			span.SetAttributes(attribute.StringSlice(k, v))
		case fmt.Stringer:
			span.SetAttributes(attribute.String(k, v.String()))
		case error:
			span.SetStatus(codes.Error, v.Error())
			span.RecordError(v)
		default:
			span.SetAttributes(attribute.String(k, spew.Sdump(v)))
		}
	}

	return nil
}
