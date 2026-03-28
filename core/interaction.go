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

// interactionKey is the rc.Values key used to store SuspendFunc.
const interactionKey = "__interaction__"

func (rc *Context) WithInteraction(plugin InteractionPlugin) {
	rc.Values[interactionKey] = plugin
}

func (rc *Context) Interaction() InteractionPlugin {
	plugin, _ := rc.Values[interactionKey].(InteractionPlugin)
	return plugin
}

type InteractionPlugin interface {
	Interact(rc *Context, i Interaction) (*InteractionResult, error)
}
