package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KKKKKKKEM/flowkit/core"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Register[Req, Resp any](r gin.IRouter, path string, app core.App[Req, Resp]) {
	r.POST(path, func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := app.Invoke(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})
}

func RegisterFunc[Req, Resp any](r gin.IRouter, path string, fn func(context.Context, Req) (Resp, error)) {
	Register(r, path, core.AppFunc[Req, Resp](fn))
}

const (
	sessionIDKey   = "SESSION-ID"
	lastEventIDKey = "Last-Event-ID"
)

type answerRequest struct {
	InteractionID string                 `json:"interaction_id"`
	Result        core.InteractionResult `json:"result"`
}

func SSE[Req, Resp any](
	r gin.IRouter,
	runPath string,
	app core.App[Req, Resp],
	store *SSESessionStore,
	buildReq func(*gin.Context) (Req, error),
	onRC ...func(*SSESession, *core.RunContext, Req),
) {
	r.POST(runPath, func(c *gin.Context) {
		lastSeq := int64(0)
		if v := c.GetHeader(lastEventIDKey); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				lastSeq = n
			}
		}

		sessionID := c.GetHeader(sessionIDKey)

		var (
			sess   *SSESession
			exists bool
		)
		if sessionID != "" {
			sess, exists = store.Get(sessionID)
		}
		if sess == nil {
			sessionID = uuid.NewString()
			sess = store.Create(sessionID)
			exists = false
		}

		ch, unsub := sess.subscribe(lastSeq)
		defer unsub()

		if !exists {
			req, err := buildReq(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			go func() {
				rc := core.NewRunContext(context.Background(), sessionID)
				rc.WithSuspend(func(i core.Interaction) (core.InteractionResult, error) {
					id := uuid.NewString()
					return sess.suspend(id, i)
				})

				for _, fn := range onRC {
					fn(sess, rc, req)
				}

				resp, err := app.Invoke(rc, req)

				sess.mu.Lock()
				sess.done = true
				sess.mu.Unlock()

				if err != nil {
					sess.emit(SSEError, gin.H{"message": err.Error()})
				} else {
					sess.emit(SSEDone, resp)
				}
			}()
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		writeSSEEvent(c, SSEEvent{Seq: 0, Type: SSEEventSession, Data: gin.H{"session_id": sessionID}})
		c.Writer.Flush()

		ctx := c.Request.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-ch:
				if !ok {
					return
				}
				writeSSEEvent(c, e)
				c.Writer.Flush()
				if e.Type == SSEDone || e.Type == SSEError {
					store.Delete(sessionID)
					return
				}
			}
		}
	})

	r.POST(runPath+"/answer", func(c *gin.Context) {
		sessionID := c.GetHeader(sessionIDKey)
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session-ID header required"})
			return
		}

		sess, ok := store.Get(sessionID)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}

		var body answerRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := sess.answer(body.InteractionID, body.Result); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}

func writeSSEEvent(c *gin.Context, e SSEEvent) {
	b, _ := json.Marshal(e.Data)
	fmt.Fprintf(c.Writer, "id: %d\nevent: %s\ndata: %s\n\n", e.Seq, e.Type, b)
}

func DefaultSSESessionStore() *SSESessionStore {
	return NewSSESessionStore(30 * time.Minute)
}
