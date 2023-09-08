package tracing

import (
	"github.com/opentracing/opentracing-go"
)

// CustomTracingHookFn is a type alias for the function that can be hooked along tracing operations.
type CustomTracingHookFn func(opentracing.Span)

var customHooks []CustomTracingHookFn

func AddCustomTracingHook(hook CustomTracingHookFn) {
	customHooks = append(customHooks, hook)
}

func RunCustomTracingHooks(span opentracing.Span) {
	for _, hook := range customHooks {
		hook(span)
	}
}
