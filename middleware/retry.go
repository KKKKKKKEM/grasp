package middleware

import "github.com/KKKKKKKEM/flowkit/core"

// RetryMiddleware 根据 ErrorPolicy 自动重试失败的 stage
func RetryMiddleware(policy core.ErrorPolicy) core.Middleware {
	return func(next core.StageRunner) core.StageRunner {
		return func(rc *core.RunContext, st core.Stage) core.StageResult {
			var result core.StageResult
			attempt := 0

			for {
				result = next(rc, st)

				if result.IsSuccess() {
					return result
				}

				if !policy.ShouldRetry(st.Name(), result.Err, attempt) {
					return result
				}

				attempt++
				// 这里可以加回退策略（如指数退避），当前简单实现不等待
			}
		}
	}
}
