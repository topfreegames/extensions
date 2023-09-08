package grpc

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/topfreegames/extensions/v9/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func WithUnaryClientTracing() grpc.DialOption {
	return grpc.WithUnaryInterceptor(OpenTracingClientInterceptor())
}

func OpenTracingClientInterceptor() grpc.UnaryClientInterceptor {
	tracer := opentracing.GlobalTracer()
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		operationName := "gRPC " + method
		span, ctx := createGrpcSpan(ctx, operationName)
		ext.SpanKindRPCClient.Set(span)
		defer span.Finish()
		defer tracing.LogPanic(span)

		ctx = injectSpanContext(ctx, tracer, span)
		err := invoker(ctx, method, req, resp, cc, opts...)
		if err != nil {
			tracing.LogError(span, err.Error())
		}
		return err
	}
}

func createGrpcSpan(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	tags := opentracing.Tags{
		string(ext.Component): "gRPC",
	}
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		childSpan := opentracing.StartSpan(operationName, opentracing.ChildOf(span.Context()), tags)
		tracing.RunCustomTracingHooks(ctx, operationName, childSpan)
		return childSpan, opentracing.ContextWithSpan(ctx, childSpan)
	}

	return opentracing.StartSpanFromContext(ctx, operationName, tags)
}

func injectSpanContext(ctx context.Context, tracer opentracing.Tracer, clientSpan opentracing.Span) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	mdWriter := metadataReaderWriter{md}
	err := tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, mdWriter)
	if err != nil {
		tracing.LogError(clientSpan, "tracer.Inject() failed: "+err.Error())
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// metadataReaderWriter satisfies both the opentracing.TextMapReader and
// opentracing.TextMapWriter interfaces.
type metadataReaderWriter struct {
	metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	// The GRPC HPACK implementation rejects any uppercase keys here.
	//
	// As such, since the HTTP_HEADERS format is case-insensitive anyway, we
	// blindly lowercase the key (which is guaranteed to work in the
	// Inject/Extract sense per the OpenTracing spec).
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
