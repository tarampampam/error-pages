package logger

import "context"

type ctxKey struct{}

// With sets the provided Logger into the context and returns the new context.
//
// Usage example:
//
//	log, _ := logger.New(logger.InfoLevel, logger.ConsoleFormat)
//	ctx = logger.With(ctx, log)
//	// or, attaching a derived logger
//	ctx = logger.With(ctx, log.Named("http"))
func With(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// WithFields adds the provided Attrs to the logger in the context and returns the new context. If
// there is no logger in the context, it returns the original one.
//
// Usage example:
//
//	ctx = logger.WithFields(ctx, logger.String("key", "value"))
//	// now context logger will have the "key" field with "value" value, if the logger is present in the context;
//	// multiple calls to WithFields will accumulate fields in the context logger
func WithFields(ctx context.Context, attrs ...Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}

	if log, ok := fromContext(ctx); ok {
		return With(ctx, log.With(attrs...))
	}

	return ctx
}

// FromContext retrieves the Logger from the context. If the logger is not found, it returns a no-op logger.
// You may use IsSet to check if the logger was explicitly set in the context before calling FromContext.
//
// The context argument is simplified to an interface with a Value method to allow using this function with
// different context implementations, not just the standard library's [context.Context].
//
// Usage example:
//
//	logger.FromContext(ctx).Info("logged if logger is present, otherwise no-op")
func FromContext(ctx interface{ Value(any) any }) *Logger {
	log, _ := fromContext(ctx)

	return log
}

// IsSet checks if a Logger is explicitly set in the context.
func IsSet(ctx interface{ Value(any) any }) bool {
	_, ok := fromContext(ctx)

	return ok
}

var noopLogger = NewNop() //nolint:gochecknoglobals

// fromContext returns the logger stored in the context and reports whether it was explicitly set. If no logger is
// found, it returns a no-op logger and false.
func fromContext(ctx interface{ Value(any) any }) (*Logger, bool) {
	if v := ctx.Value(ctxKey{}); v != nil {
		if log, ok := v.(*Logger); ok {
			return log, true
		}
	}

	return noopLogger, false
}
