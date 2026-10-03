package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SelectAgents prompts the user to select which agent(s) to activate in a preset
func SelectAgents(presetName string, availableAgents []string) ([]string, error) {
	if len(availableAgents) == 0 {
		return nil, fmt.Errorf("no agents found in preset %q", presetName)
	}

	// Fast path: single agent, automatically select
	if len(availableAgents) == 1 {
		return availableAgents, nil
	}

	// Check if running in a non-interactive pipe or test
	fileInfo, err := os.Stdin.Stat()
	if err != nil || (fileInfo.Mode()&os.ModeCharDevice) == 0 {
		// Non-interactive: activate all available agents
		return availableAgents, nil
	}

	fmt.Printf("\nPreset %s contains multiple agents. Select which to activate:\n\n", Bold(Cyan(presetName)))
	for i, ag := range availableAgents {
		fmt.Printf("  [%d] %s\n", i+1, Green(ag))
	}
	fmt.Printf("  [%d] %s\n", len(availableAgents)+1, Bold("All of the above (recommended)"))
	fmt.Print("\nEnter choice (number or comma-separated, default: All): ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return availableAgents, nil
	}

	input := strings.TrimSpace(line)
	if input == "" {
		return availableAgents, nil
	}

	allOption := len(availableAgents) + 1
	var selected []string
	parts := strings.Split(input, ",")
	for _, p := range parts {
		idx, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			continue
		}
		if idx == allOption {
			return availableAgents, nil
		}
		if idx >= 1 && idx <= len(availableAgents) {
			selected = append(selected, availableAgents[idx-1])
		}
	}

	if len(selected) == 0 {
		return availableAgents, nil
	}

	return selected, nil
}
