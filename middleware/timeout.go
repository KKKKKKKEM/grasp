package middleware

import (
	"fmt"
	"time"

	"github.com/KKKKKKKEM/flowkit/core"
)

// TimeoutMiddleware 给 stage 增加超时限制（可选）
// 如果 stage 执行超过指定时间则返回失败
func TimeoutMiddleware(timeout time.Duration) core.Middleware {
	return func(next core.StageRunner) core.StageRunner {
		return func(rc *core.Context, st core.Stage) core.StageResult {
			// 创建带超时的子上下文
			childRc, cancel := rc.WithCancel()
			defer cancel()

			// 用 done 通道来等待执行完成
			done := make(chan core.StageResult, 1)
			go func() {
				done <- next(childRc, st)
			}()

			timer := time.NewTimer(timeout)
			defer timer.Stop()

			select {
			case result := <-done:
				return result
			case <-timer.C:
				cancel()
				return core.StageResult{
					Status: core.StageFailed,
					Err:    fmt.Errorf("stage timeout after %v", timeout),
				}
			}
		}
	}
}
