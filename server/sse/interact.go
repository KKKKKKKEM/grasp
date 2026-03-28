package sse

import (
	"encoding/json"
	"fmt"

	"github.com/KKKKKKKEM/flowkit/core"
	"github.com/google/uuid"
)

type InteractionPlugin struct {
	session *Session
}

func (p *InteractionPlugin) FormatResult(_ *core.Context, _ core.Interaction, result *core.InteractionResult) (*core.InteractionResult, error) {
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", result.Answer)), &result.Answer)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *InteractionPlugin) Interact(_ *core.Context, i core.Interaction) (*core.InteractionResult, error) {
	interactionID := uuid.NewString()
	return p.session.suspend(interactionID, i)
}

func NewSSEInteractionPlugin(session *Session) *InteractionPlugin {
	return &InteractionPlugin{session: session}
}
