package middleware

import "github.com/KKKKKKKEM/flowkit/core"

// MetricsMiddleware 统计 stage 的执行指标（调用次数、失败次数等）
// 这里简单示例，实现中可以集成到监控系统
func MetricsMiddleware(metricsCollector map[string]map[string]float64) core.Middleware {
	return func(next core.StageRunner) core.StageRunner {
		return func(rc *core.Context, st core.Stage) core.StageResult {
			stageName := st.Name()
			if metricsCollector[stageName] == nil {
				metricsCollector[stageName] = make(map[string]float64)
			}

			metrics := metricsCollector[stageName]
			metrics["total_calls"]++

			result := next(rc, st)

			if result.IsFailed() {
				metrics["total_failures"]++
			} else if result.Status == core.StageSuccess {
				metrics["total_success"]++
			}

			return result
		}
	}
}
