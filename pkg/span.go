package pkg

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func StartSpanWithContext(ctx context.Context, tracer opentracing.Tracer, operationName string) (context.Context, opentracing.Span) {
	spanOptions := make([]opentracing.StartSpanOption, 0, 1)
	if parentSpanContext, ok := ctx.Value(TracerParentCtxValue).(opentracing.Span); ok {
		spanOptions = append(spanOptions, opentracing.ChildOf(parentSpanContext.Context()))
	}

	span := tracer.StartSpan(operationName, spanOptions...)
	return context.WithValue(ctx, TracerParentCtxValue, span), span
}

func SpanError(span opentracing.Span, err error) error {
	span.SetTag("error", true)
	span.LogFields(log.String("error", err.Error()))
	return err
}
