package core

import (
	"context"
	"sync"
	"time"
)

type SharedState map[string]any

type sharedValues struct {
	mu   sync.RWMutex
	data SharedState
}

// Context 实现 context.Context 接口，同时承载业务数据
// 这是整个框架的核心上下文对象，所有 Stage 通过它进行数据共享与信号传递
type Context struct {
	ctx       context.Context
	TraceID   string
	StartedAt time.Time

	state *sharedValues
	Tags  map[string]string
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
		if v, exist := rc.Get(k); exist {
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
		state:     &sharedValues{data: make(SharedState)},
		Tags:      make(map[string]string),
		StartedAt: time.Now(),
	}
}

func (rc *Context) WithContext(ctx context.Context) *Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Context{
		ctx:       ctx,
		TraceID:   rc.TraceID,
		state:     rc.state,
		Tags:      rc.Tags,
		StartedAt: rc.StartedAt,
	}
}

// WithTimeout 返回一个新的带超时的 Context
func (rc *Context) WithTimeout(d time.Duration) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(rc.ctx, d)
	return rc.WithContext(ctx), cancel
}

// WithCancel 返回一个新的可取消的 Context
func (rc *Context) WithCancel() (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(rc.ctx)
	return rc.WithContext(ctx), cancel
}

// WithValue 返回一个新的包含值的 Context
func (rc *Context) WithValue(key string, val any) *Context {
	rc.Set(key, val)
	return rc
}

func (rc *Context) Set(key string, val any) {
	rc.state.mu.Lock()
	defer rc.state.mu.Unlock()
	rc.state.data[key] = val
}

func (rc *Context) Get(key string) (any, bool) {
	rc.state.mu.RLock()
	defer rc.state.mu.RUnlock()
	v, ok := rc.state.data[key]
	return v, ok
}

func (rc *Context) Merge(values map[string]any) {
	if len(values) == 0 {
		return
	}
	rc.state.mu.Lock()
	defer rc.state.mu.Unlock()
	for k, v := range values {
		rc.state.data[k] = v
	}
}

func (rc *Context) Fork(traceID string) *Context {
	if traceID == "" {
		traceID = rc.TraceID
	}
	return &Context{
		ctx:       rc,
		TraceID:   traceID,
		state:     &sharedValues{data: make(SharedState)},
		Tags:      make(map[string]string),
		StartedAt: rc.StartedAt,
	}
}

func (rc *Context) Derive(traceID string) *Context {
	child := rc.WithContext(rc)
	if traceID != "" {
		child.TraceID = traceID
	}
	return child
}

// Duration 返回从启动到现在的耗时
func (rc *Context) Duration() time.Duration {
	return time.Since(rc.StartedAt)
}
