package server

import (
	"github.com/KKKKKKKEM/flowkit/core"
	"github.com/gin-gonic/gin"
)

func Func[Req, Resp any](fn func(*core.Context, Req) (Resp, error)) core.App[Req, Resp] {
	return core.AppFunc[Req, Resp](fn)
}

func DefaultBuildReq[Req any](c *gin.Context) (Req, error) {
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		return req, err
	}
	return req, nil
}
