package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// CustomTracingHookFn is a type alias for the function that can be hooked along tracing operations.
type CustomTracingHookFn func(context.Context, string, opentracing.Span)

var customHooks []CustomTracingHookFn

func AddCustomTracingHook(hook CustomTracingHookFn) {
	customHooks = append(customHooks, hook)
}

func RunCustomTracingHooks(ctx context.Context, operationName string, span opentracing.Span) {
	for _, hook := range customHooks {
		hook(ctx, operationName, span)
	}
}

type CustomTracingTagsHookFn func(context.Context, opentracing.Tags) opentracing.Tags

var customTagsHooks []CustomTracingTagsHookFn

func AddCustomTracingTagsHook(hook CustomTracingTagsHookFn) {
	customTagsHooks = append(customTagsHooks, hook)
}

func RunCustomTracingTagsHooks(ctx context.Context, tags opentracing.Tags) opentracing.Tags {
	for _, hook := range customTagsHooks {
		tags = hook(ctx, tags)
	}
	return tags
}
