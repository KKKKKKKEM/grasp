package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/KKKKKKKEM/flowkit/core"
)

// LoggingMiddleware 记录 stage 的执行（开始、结束、耗时、错误）
func LoggingMiddleware(logger *log.Logger) core.Middleware {
	return func(next core.StageRunner) core.StageRunner {
		return func(rc *core.RunContext, st core.Stage) core.StageResult {
			nameTag := fmt.Sprintf("[%s]", st.Name())
			logger.Printf("%s Stage started (traceID=%s)", nameTag, rc.TraceID)

			start := time.Now()
			result := next(rc, st)
			elapsed := time.Since(start)

			if result.IsFailed() {
				logger.Printf("%s Stage failed after %v. Err: %v", nameTag, elapsed, result.Err)
			} else if result.Status == core.StageSkipped {
				logger.Printf("%s Stage skipped", nameTag)
			} else {
				logger.Printf("%s Stage succeeded in %v", nameTag, elapsed)
			}

			if result.Metrics == nil {
				result.Metrics = make(map[string]float64)
			}
			result.Metrics["duration_ms"] = float64(elapsed.Milliseconds())

			return result
		}
	}
}
