package yredis

import (
	"context"
	"fmt"

	"github.com/AmarsDing/lib"
	"github.com/AmarsDing/lib/internal/json"
	"github.com/AmarsDing/lib/net/ytrace"
	"github.com/AmarsDing/lib/os/ycmd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// tracingItem holds the information for redis tracing.
type tracingItem struct {
	err         error
	commandName string
	arguments   []interface{}
	costMilli   int64
}

const (
	tracingInstrumentName               = "github.com/AmarsDing/lib/database/yredis"
	tracingAttrRedisHost                = "redis.host"
	tracingAttrRedisPort                = "redis.port"
	tracingAttrRedisDb                  = "redis.db"
	tracingEventRedisExecution          = "redis.execution"
	tracingEventRedisExecutionCommand   = "redis.execution.command"
	tracingEventRedisExecutionCost      = "redis.execution.cost"
	tracingEventRedisExecutionArguments = "redis.execution.arguments"
)

var (
	// tracingInternal enables tracing for internal type spans.
	// It's true in default.
	tracingInternal = true
)

func init() {
	tracingInternal = ycmd.GetOptWithEnv("lib.tracing.internal", true).Bool()
}

// addTracingItem checks and adds redis tracing information to OpenTelemetry.
func (c *Conn) addTracingItem(item *tracingItem) {
	if !tracingInternal || !ytrace.IsActivated(c.ctx) {
		return
	}
	tr := otel.GetTracerProvider().Tracer(
		tracingInstrumentName,
		trace.WithInstrumentationVersion(lib.VERSION),
	)
	ctx := c.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	_, span := tr.Start(ctx, "Redis."+item.commandName, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	if item.err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf(`%+v`, item.err))
	}
	span.SetAttributes(ytrace.CommonLabels()...)
	span.SetAttributes(
		attribute.String(tracingAttrRedisHost, c.redis.config.Host),
		attribute.Int(tracingAttrRedisPort, c.redis.config.Port),
		attribute.Int(tracingAttrRedisDb, c.redis.config.Db),
	)
	jsonBytes, _ := json.Marshal(item.arguments)
	span.AddEvent(tracingEventRedisExecution, trace.WithAttributes(
		attribute.String(tracingEventRedisExecutionCommand, item.commandName),
		attribute.String(tracingEventRedisExecutionCost, fmt.Sprintf(`%d ms`, item.costMilli)),
		attribute.String(tracingEventRedisExecutionArguments, string(jsonBytes)),
	))
}
