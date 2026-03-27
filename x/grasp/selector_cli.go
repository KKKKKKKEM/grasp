package grasp

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/KKKKKKKEM/flowkit/builtin/extract"
)

func InteractiveCLI() SelectFunc {
	return func(_ context.Context, items []extract.ParseItem) ([]extract.ParseItem, error) {
		if len(items) == 0 {
			return nil, nil
		}

		fmt.Println()
		for i, item := range items {
			fmt.Printf("  [%d] %s\n", i, item.Name)
			if item.URI != "" {
				fmt.Printf("      %s\n", item.URI)
			}
		}
		fmt.Println()
		fmt.Printf("Select items (e.g. 0,2 / 0-3 / all) [default: all]: ")

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}
		line = strings.TrimSpace(line)

		if line == "" || line == "all" {
			return items, nil
		}

		indices, err := parseIndices(line, len(items))
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q: %w", line, err)
		}

		var selected []extract.ParseItem
		for _, idx := range indices {
			selected = append(selected, items[idx])
		}
		return selected, nil
	}
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
