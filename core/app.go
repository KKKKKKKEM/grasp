package core

import "context"

// App 是框架中"可被外部调用的执行单元"的最小契约。
//
// Req 是传输层可序列化的请求类型。
// Resp 是传输层可序列化的响应类型。
//
// 实现者负责将 Req 翻译成内部执行所需的领域对象、执行核心逻辑、再将结果翻译成 Resp。
// 框架的 serve 层负责从传输层读取 bytes → 反序列化成 Req，再将 Resp 序列化写回传输层。
type App[Req, Resp any] interface {
	Invoke(ctx context.Context, req Req) (Resp, error)
}

// AppFunc 是 App 的函数类型实现，让任意符合签名的函数直接满足 App 接口，无需定义新类型。
type AppFunc[Req, Resp any] func(context.Context, Req) (Resp, error)

func (f AppFunc[Req, Resp]) Invoke(ctx context.Context, req Req) (Resp, error) {
	return f(ctx, req)
}
