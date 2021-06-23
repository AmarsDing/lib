package ytrace

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/container/yvar"
	"github.com/AmarsDing/lib/net/yipv4"
	"github.com/AmarsDing/lib/os/ycmd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracingCommonKeyIpIntranet = `ip.intranet`
	tracingCommonKeyIpHostname = `hostname`
	cmdEnvKey                  = "gf.gtrace" // Configuration key for command argument or environment.
)

var (
	intranetIps, _           = yipv4.GetIntranetIpArray()
	intranetIpStr            = strings.Join(intranetIps, ",")
	hostname, _              = os.Hostname()
	tracingMaxContentLogSize = 256 * 1024 // Max log size for request and response body, especially for HTTP/RPC request.
	// defaultTextMapPropagator is the default propagator for context propagation between peers.
	defaultTextMapPropagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
)

func init() {
	if maxContentLogSize := ycmd.GetOptWithEnv(fmt.Sprintf("%s.maxcontentlogsize", cmdEnvKey)).Int(); maxContentLogSize > 0 {
		tracingMaxContentLogSize = maxContentLogSize
	}
	CheckSetDefaultTextMapPropagator()
}

// MaxContentLogSize returns the max log size for request and response body, especially for HTTP/RPC request.
func MaxContentLogSize() int {
	return tracingMaxContentLogSize
}

// CommonLabels returns common used attribute labels:
// ip.intranet, hostname.
func CommonLabels() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(tracingCommonKeyIpHostname, hostname),
		attribute.String(tracingCommonKeyIpIntranet, intranetIpStr),
	}
}

// IsActivated checks and returns if tracing feature is activated.
func IsActivated(ctx context.Context) bool {
	return GetTraceId(ctx) != ""
}

// CheckSetDefaultTextMapPropagator sets the default TextMapPropagator if it is not set previously.
func CheckSetDefaultTextMapPropagator() {
	p := otel.GetTextMapPropagator()
	if len(p.Fields()) == 0 {
		otel.SetTextMapPropagator(GetDefaultTextMapPropagator())
	}
}

// GetDefaultTextMapPropagator returns the default propagator for context propagation between peers.
func GetDefaultTextMapPropagator() propagation.TextMapPropagator {
	return defaultTextMapPropagator
}

// GetTraceId retrieves and returns TraceId from context.
// It returns an empty string is tracing feature is not activated.
func GetTraceId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceId := trace.SpanContextFromContext(ctx).TraceID()
	if traceId.IsValid() {
		return traceId.String()
	}
	return ""
}

// GetSpanId retrieves and returns SpanId from context.
// It returns an empty string is tracing feature is not activated.
func GetSpanId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	spanId := trace.SpanContextFromContext(ctx).SpanID()
	if spanId.IsValid() {
		return spanId.String()
	}
	return ""
}

// SetBaggageValue is a convenient function for adding one key-value pair to baggage.
// Note that it uses attribute.Any to set the key-value pair.
func SetBaggageValue(ctx context.Context, key string, value interface{}) context.Context {
	return NewBaggage(ctx).SetValue(key, value)
}

// SetBaggageMap is a convenient function for adding map key-value pairs to baggage.
// Note that it uses attribute.Any to set the key-value pair.
func SetBaggageMap(ctx context.Context, data map[string]interface{}) context.Context {
	return NewBaggage(ctx).SetMap(data)
}

// GetBaggageMap retrieves and returns the baggage values as map.
func GetBaggageMap(ctx context.Context) *ymap.StrAnyMap {
	return NewBaggage(ctx).GetMap()
}

// GetBaggageVar retrieves value and returns a *yvar.Var for specified key from baggage.
func GetBaggageVar(ctx context.Context, key string) *yvar.Var {
	return NewBaggage(ctx).GetVar(key)
}
