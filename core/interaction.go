package core

import "fmt"

type InteractionType string

type Interaction struct {
	Type    InteractionType `json:"type"`
	Payload any             `json:"payload"`
}

type InteractionResult struct {
	Answer any `json:"answer"`
}

type ErrInteractionRequired struct {
	Interaction Interaction
}

func (e *ErrInteractionRequired) Error() string {
	return fmt.Sprintf("interaction required: type=%s", e.Interaction.Type)
}

// SuspendFunc is injected into Context by the SSE framework layer.
// When a plugin calls it, the pipeline suspends: the Interaction is pushed
// to the SSE stream and the call blocks until the client submits an answer
// via POST /answer, at which point the result is returned and execution resumes.
//
// CLI plugins ignore this and block on stdin instead.
type SuspendFunc func(i Interaction) (InteractionResult, error)

// suspendKey is the rc.Values key used to store SuspendFunc.
const suspendKey = "__suspend__"

// WithSuspend injects a SuspendFunc into the Context (called by the framework).
func (rc *Context) WithSuspend(fn SuspendFunc) {
	rc.Values[suspendKey] = fn
}

// Suspend returns the SuspendFunc if one was injected (Web/SSE mode), or nil (CLI mode).
func (rc *Context) Suspend() SuspendFunc {
	fn, _ := rc.Values[suspendKey].(SuspendFunc)
	return fn
}

// InteractionPlugin handles a specific interaction type.
//
// CLI mode: Interact blocks on stdin (or any other synchronous source).
// Web/SSE mode: Interact calls rc.Suspend()(i) which blocks until the client
// submits an answer through the SSE channel; the framework injects SuspendFunc
// into rc before running the pipeline.
//
// The interface is identical in both modes — only the SuspendFunc injected by
// the framework differs.
type InteractionPlugin interface {
	Type() InteractionType
	Interact(rc *Context, i Interaction) error
}
