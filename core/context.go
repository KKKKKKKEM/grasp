package core

import (
	"context"
	"time"
)

type SharedState map[string]any

// Context 实现 context.Context 接口，同时承载业务数据
// 这是整个框架的核心上下文对象，所有 Stage 通过它进行数据共享与信号传递
type Context struct {
	ctx       context.Context
	TraceID   string
	Values    SharedState // 执行过程中的中间产物与共享数据
	Tags      map[string]string
	StartedAt time.Time
}

func (rc *Context) Deadline() (deadline time.Time, ok bool) {
	return rc.ctx.Deadline()
}

func (rc *Context) Done() <-chan struct{} {
	return rc.ctx.Done()
}

func (rc *Context) Err() error {
	return rc.ctx.Err()
}

// Value 先查业务 Values，再落回到底层 context
func (rc *Context) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok {
		if v, exist := rc.Values[k]; exist {
			return v
		}
	}
	return rc.ctx.Value(key)
}

// NewContext 创建一个新的 Context
func NewContext(ctx context.Context, traceID string) *Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Context{
		ctx:       ctx,
		TraceID:   traceID,
		Values:    make(SharedState),
		Tags:      make(map[string]string),
		StartedAt: time.Now(),
	}
}

// WithTimeout 返回一个新的带超时的 Context
func (rc *Context) WithTimeout(d time.Duration) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(rc.ctx, d)
	return &Context{
		ctx:       ctx,
		TraceID:   rc.TraceID,
		Values:    rc.Values,
		Tags:      rc.Tags,
		StartedAt: rc.StartedAt,
	}, cancel
}

// WithCancel 返回一个新的可取消的 Context
func (rc *Context) WithCancel() (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(rc.ctx)
	return &Context{
		ctx:       ctx,
		TraceID:   rc.TraceID,
		Values:    rc.Values,
		Tags:      rc.Tags,
		StartedAt: rc.StartedAt,
	}, cancel
}

// WithValue 返回一个新的包含值的 Context
func (rc *Context) WithValue(key string, val any) *Context {
	rc.Values[key] = val
	return rc
}

// Duration 返回从启动到现在的耗时
func (rc *Context) Duration() time.Duration {
	return time.Since(rc.StartedAt)
}
