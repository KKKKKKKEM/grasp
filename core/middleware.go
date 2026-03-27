package core

// StageRunner 是运行单个 stage 的函数签名
type StageRunner func(rc *RunContext, st Stage) StageResult

// Middleware 是中间件函数，接收一个 StageRunner 返回一个包装过的 StageRunner
type Middleware func(next StageRunner) StageRunner

// Chain 将多个中间件链接起来
func Chain(mws ...Middleware) Middleware {
	return func(final StageRunner) StageRunner {
		wrapped := final
		for i := len(mws) - 1; i >= 0; i-- {
			wrapped = mws[i](wrapped)
		}
		return wrapped
	}
}
