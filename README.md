# Example usage

```go
package main 

import (
    "errors"
    "context"
    "os"
    "strconv"

    "github.com/sirupsen/logrus"
)

func main() {
    ctx := context.TODO()

    // Configure OTEL

    debug, err := strconv.ParseBool(os.Getenv("OTEL_EXPORTER_DEBUG_MODE"))
    if err != nil {
        // ...
    }

    tp, err := logrusotel.NewTracerProvider(
        os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT"),
        "provider name",
        debug,
    )
    if err != nil {
        // ...
    }

    defer tp.Shutdown(ctx)

    tracer := tp.Tracer("tracer name")

    // Configure logrus hook

    logrus.AddHook(logrusotel.NewJaegerHook())

    // Usage

    ctx, span := tracer.Start(ctx, "my operation")
    defer span.End()

    logrus.WithContext(ctx).
        WithField("param", "value").
        Info("some log")
    
    logrus.WithContext(ctx).
        WithError(errors.New("some error")).
        Error("another log")
}
```