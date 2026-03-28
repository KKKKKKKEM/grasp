package grasp

import (
	"github.com/KKKKKKKEM/flowkit/server"
	"github.com/KKKKKKKEM/flowkit/server/sse"
	"github.com/gin-gonic/gin"
)

func (p *Pipeline) GinRegister(r gin.IRouter) gin.IRouter {
	sse.SSE(r, "/grasp", sse.Config[*Task, *Report]{
		App: server.Func(p.Invoke),
	})
	return r
}

func (p *Pipeline) Serve(addr string) error {
	engine := gin.Default()
	p.GinRegister(engine)
	return engine.Run(addr)
}
