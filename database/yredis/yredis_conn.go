package yredis

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/AmarsDing/lib/container/yvar"
	"github.com/AmarsDing/lib/internal/json"
	"github.com/AmarsDing/lib/os/ytime"
	"github.com/AmarsDing/lib/util/yconv"
	"github.com/gomodule/redigo/redis"
)

// Do sends a command to the server and returns the received reply.
// It uses json.Marshal for struct/slice/map type values before committing them to redis.
// The timeout overrides the read timeout set when dialing the connection.
func (c *Conn) do(timeout time.Duration, commandName string, args ...interface{}) (reply interface{}, err error) {
	var (
		reflectValue reflect.Value
		reflectKind  reflect.Kind
	)
	for k, v := range args {
		reflectValue = reflect.ValueOf(v)
		reflectKind = reflectValue.Kind()
		if reflectKind == reflect.Ptr {
			reflectValue = reflectValue.Elem()
			reflectKind = reflectValue.Kind()
		}
		switch reflectKind {
		case
			reflect.Struct,
			reflect.Map,
			reflect.Slice,
			reflect.Array:
			// Ignore slice type of: []byte.
			if _, ok := v.([]byte); !ok {
				if args[k], err = json.Marshal(v); err != nil {
					return nil, err
				}
			}
		}
	}
	if timeout > 0 {
		conn, ok := c.Conn.(redis.ConnWithTimeout)
		if !ok {
			return yvar.New(nil), errors.New(`current connection does not support "ConnWithTimeout"`)
		}
		return conn.DoWithTimeout(timeout, commandName, args...)
	}
	timestampMilli1 := ytime.TimestampMilli()
	reply, err = c.Conn.Do(commandName, args...)
	timestampMilli2 := ytime.TimestampMilli()

	// Tracing.
	c.addTracingItem(&tracingItem{
		err:         err,
		commandName: commandName,
		arguments:   args,
		costMilli:   timestampMilli2 - timestampMilli1,
	})
	return
}

// Ctx is a channing function which sets the context for next operation.
func (c *Conn) Ctx(ctx context.Context) *Conn {
	c.ctx = ctx
	return c
}

// Do sends a command to the server and returns the received reply.
// It uses json.Marshal for struct/slice/map type values before committing them to redis.
func (c *Conn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	return c.do(0, commandName, args...)
}

// DoWithTimeout sends a command to the server and returns the received reply.
// The timeout overrides the read timeout set when dialing the connection.
func (c *Conn) DoWithTimeout(timeout time.Duration, commandName string, args ...interface{}) (reply interface{}, err error) {
	return c.do(timeout, commandName, args...)
}

// DoVar retrieves and returns the result from command as yvar.Var.
func (c *Conn) DoVar(commandName string, args ...interface{}) (*yvar.Var, error) {
	return resultToVar(c.Do(commandName, args...))
}

// DoVarWithTimeout retrieves and returns the result from command as yvar.Var.
// The timeout overrides the read timeout set when dialing the connection.
func (c *Conn) DoVarWithTimeout(timeout time.Duration, commandName string, args ...interface{}) (*yvar.Var, error) {
	return resultToVar(c.DoWithTimeout(timeout, commandName, args...))
}

// ReceiveVar receives a single reply as yvar.Var from the Redis server.
func (c *Conn) ReceiveVar() (*yvar.Var, error) {
	return resultToVar(c.Receive())
}

// ReceiveVarWithTimeout receives a single reply as yvar.Var from the Redis server.
// The timeout overrides the read timeout set when dialing the connection.
func (c *Conn) ReceiveVarWithTimeout(timeout time.Duration) (*yvar.Var, error) {
	conn, ok := c.Conn.(redis.ConnWithTimeout)
	if !ok {
		return yvar.New(nil), errors.New(`current connection does not support "ConnWithTimeout"`)
	}
	return resultToVar(conn.ReceiveWithTimeout(timeout))
}

// resultToVar converts redis operation result to yvar.Var.
func resultToVar(result interface{}, err error) (*yvar.Var, error) {
	if err == nil {
		if result, ok := result.([]byte); ok {
			return yvar.New(yconv.UnsafeBytesToStr(result)), err
		}
		// It treats all returned slice as string slice.
		if result, ok := result.([]interface{}); ok {
			return yvar.New(yconv.Strings(result)), err
		}
	}
	return yvar.New(result), err
}
