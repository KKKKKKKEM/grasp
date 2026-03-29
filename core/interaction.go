package core

import "fmt"

type InteractionType string

const (
	InteractionTypeUserInput InteractionType = "user_input"
	InteractionTypeApproval  InteractionType = "approval"
	InteractionTypeCaptcha   InteractionType = "captcha"
	InteractionTypeCustom    InteractionType = "custom"
	InteractionTypeSelect    InteractionType = "select"
)

type Interaction struct {
	Type    InteractionType `json:"type"`
	Payload any             `json:"payload"`
	Message string          `json:"message,omitempty"`
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

func (rc *Context) WithInteractionPlugin(plugin InteractionPlugin) {
	rc.Set(interactionKey, plugin)
}

func (rc *Context) InteractionPlugin() InteractionPlugin {
	v, _ := rc.Get(interactionKey)
	plugin, _ := v.(InteractionPlugin)
	return plugin
}

type InteractionPlugin interface {
	Interact(rc *Context, i Interaction) (*InteractionResult, error)
	FormatResult(rc *Context, i Interaction, result *InteractionResult) (*InteractionResult, error)
}
