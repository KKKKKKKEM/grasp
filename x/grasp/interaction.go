package grasp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/KKKKKKKEM/flowkit/builtin/extract"
	"github.com/KKKKKKKEM/flowkit/core"
)

const InteractionTypeSelect core.InteractionType = "select"

// CLISelectPlugin handles item selection by blocking on stdin.
type CLISelectPlugin struct{}

func (CLISelectPlugin) Type() core.InteractionType { return InteractionTypeSelect }

func (CLISelectPlugin) Interact(rc *core.RunContext, i core.Interaction) error {
	items, err := payloadToItems(i.Payload)
	if err != nil {
		return fmt.Errorf("select interact: bad payload: %w", err)
	}

	fmt.Println()
	for idx, item := range items {
		fmt.Printf("  [%d] %s\n", idx, item.Name)
		if item.URI != "" {
			fmt.Printf("      %s\n", item.URI)
		}
	}
	fmt.Println()
	fmt.Printf("Select items (e.g. 0,2 / 0-3 / all) [default: all]: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("select interact: read input: %w", err)
	}
	line = strings.TrimSpace(line)

	var selected []extract.ParseItem
	if line == "" || line == "all" {
		selected = items
	} else {
		indices, err := parseIndices(line, len(items))
		if err != nil {
			return fmt.Errorf("select interact: invalid selection %q: %w", line, err)
		}
		for _, idx := range indices {
			selected = append(selected, items[idx])
		}
	}

	rc.WithValue("selected_items", selected)
	return nil
}

// WebSelectPlugin handles item selection via the SSE suspend/resume protocol.
// rc must have a SuspendFunc injected by the framework (via rc.WithSuspend).
type WebSelectPlugin struct{}

func (WebSelectPlugin) Type() core.InteractionType { return InteractionTypeSelect }

func (WebSelectPlugin) Interact(rc *core.RunContext, i core.Interaction) error {
	suspend := rc.Suspend()
	if suspend == nil {
		return fmt.Errorf("WebSelectPlugin: no SuspendFunc in rc; use CLISelectPlugin for CLI mode")
	}

	result, err := suspend(i)
	if err != nil {
		return fmt.Errorf("select interact: suspend: %w", err)
	}

	items, err := payloadToItems(i.Payload)
	if err != nil {
		return fmt.Errorf("select interact: bad payload: %w", err)
	}

	selected, err := applySelectAnswer(result, items)
	if err != nil {
		return fmt.Errorf("select interact: apply answer: %w", err)
	}

	rc.WithValue("selected_items", selected)
	return nil
}

func applySelectAnswer(result core.InteractionResult, items []extract.ParseItem) ([]extract.ParseItem, error) {
	b, err := json.Marshal(result.Answer)
	if err != nil {
		return nil, fmt.Errorf("marshal answer: %w", err)
	}
	var indices []int
	if err := json.Unmarshal(b, &indices); err != nil {
		return nil, fmt.Errorf("answer must be []int (indices): %w", err)
	}
	var selected []extract.ParseItem
	for _, idx := range indices {
		if idx < 0 || idx >= len(items) {
			return nil, fmt.Errorf("index %d out of range", idx)
		}
		selected = append(selected, items[idx])
	}
	return selected, nil
}

func payloadToItems(payload any) ([]extract.ParseItem, error) {
	if items, ok := payload.([]extract.ParseItem); ok {
		return items, nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var items []extract.ParseItem
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func parseIndices(input string, total int) ([]int, error) {
	seen := make(map[int]struct{})
	var out []int

	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			lo, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			hi, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err1 != nil || err2 != nil || lo > hi {
				return nil, fmt.Errorf("invalid range %q", part)
			}
			for i := lo; i <= hi; i++ {
				if i < 0 || i >= total {
					return nil, fmt.Errorf("index %d out of range [0, %d)", i, total)
				}
				if _, exists := seen[i]; !exists {
					seen[i] = struct{}{}
					out = append(out, i)
				}
			}
		} else {
			idx, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid index %q", part)
			}
			if idx < 0 || idx >= total {
				return nil, fmt.Errorf("index %d out of range [0, %d)", idx, total)
			}
			if _, exists := seen[idx]; !exists {
				seen[idx] = struct{}{}
				out = append(out, idx)
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no valid indices")
	}
	return out, nil
}
