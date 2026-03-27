package middleware

import (
	"fmt"
	"log"

	"github.com/KKKKKKKEM/flowkit/core"
)

// RecoveryMiddleware 防止 panic，将其转换为 StageFailed 错误
func RecoveryMiddleware(logger *log.Logger) core.Middleware {
	return func(next core.StageRunner) core.StageRunner {
		return func(rc *core.RunContext, st core.Stage) (result core.StageResult) {
			defer func() {
				if r := recover(); r != nil {
					logger.Printf("[%s] Stage panicked: %v", st.Name(), r)
					result = core.StageResult{
						Status: core.StageFailed,
						Err:    fmt.Errorf("stage panic: %v", r),
					}
				}
			}()

			result = next(rc, st)
			return result
		}
	}
}
