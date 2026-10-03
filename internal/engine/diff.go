package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agentenv/internal/config"
)

// DiffResult contains unified diff output and whether differences were found
type DiffResult struct {
	HasDiff bool
	Output  string
}

// GenerateUnifiedDiff generates a standard unified diff between two strings
func GenerateUnifiedDiff(src, dst, srcLabel, dstLabel string) DiffResult {
	srcLines := strings.Split(src, "\n")
	dstLines := strings.Split(dst, "\n")

	// If contents are identical, no diff
	if src == dst {
		return DiffResult{HasDiff: false, Output: ""}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s\n", srcLabel))
	sb.WriteString(fmt.Sprintf("+++ %s\n", dstLabel))

	// Simple LCS-based or line-matching diff for unified output
	diffOps := computeDiff(srcLines, dstLines)

	inHunk := false
	for _, op := range diffOps {
		if !inHunk {
			sb.WriteString("@@ -1 +1 @@\n")
			inHunk = true
		}
		sb.WriteString(op + "\n")
	}

	return DiffResult{
		HasDiff: true,
		Output:  sb.String(),
	}
}

type diffOp struct {
	op   rune // ' ', '+', '-'
	text string
}

func computeDiff(a, b []string) []string {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if a[i] == b[j] {
				lcs[i+1][j+1] = lcs[i][j] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i+1][j+1] = lcs[i+1][j]
			} else {
				lcs[i+1][j+1] = lcs[i][j+1]
			}
		}
	}

	var ops []string
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			ops = append([]string{" " + a[i-1]}, ops...)
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			ops = append([]string{"+" + b[j-1]}, ops...)
			j--
		} else if i > 0 && (j == 0 || lcs[i][j-1] < lcs[i-1][j]) {
			ops = append([]string{"-" + a[i-1]}, ops...)
			i--
		}
	}

	return ops
}

// PresetDiff compares all override files in a preset against their corresponding host files
func PresetDiff(cfg *config.Config, presetName string) (string, error) {
	if err := config.ValidatePresetName(presetName); err != nil {
		return "", err
	}

	presetDir, err := config.PresetDir(presetName)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(presetDir); os.IsNotExist(err) {
		return "", fmt.Errorf("preset %q does not exist in %s", presetName, presetDir)
	}

	var sb strings.Builder
	hasAnyDiff := false

	// Compare shared items against each agent base
	sharedDir := filepath.Join(presetDir, "shared")
	if info, err := os.Stat(sharedDir); err == nil && info.IsDir() {
		// List files in shared
		_ = filepath.Walk(sharedDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(sharedDir, path)
			// Compare with all agents
			for agName, agCfg := range cfg.Agents {
				hostBase, _ := agCfg.ResolvedHostDir()
				hostFile := filepath.Join(hostBase, rel)

				diff := diffFiles(hostFile, path, fmt.Sprintf("~/.%s/%s", agName, rel), fmt.Sprintf("presets/%s/shared/%s", presetName, rel))
				if diff.HasDiff {
					hasAnyDiff = true
					sb.WriteString(diff.Output + "\n")
				}
			}
			return nil
		})
	}

	// Compare agent-specific items
	for agName, agCfg := range cfg.Agents {
		agDir := filepath.Join(presetDir, agName)
		if info, err := os.Stat(agDir); err == nil && info.IsDir() {
			hostBase, _ := agCfg.ResolvedHostDir()

			_ = filepath.Walk(agDir, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(agDir, path)
				hostFile := filepath.Join(hostBase, rel)

				diff := diffFiles(hostFile, path, fmt.Sprintf("~/.%s/%s", agName, rel), fmt.Sprintf("presets/%s/%s/%s", presetName, agName, rel))
				if diff.HasDiff {
					hasAnyDiff = true
					sb.WriteString(diff.Output + "\n")
				}
				return nil
			})
		}
	}

	if !hasAnyDiff {
		return "No configuration differences detected between preset and host baseline.\n", nil
	}

	return sb.String(), nil
}

func diffFiles(hostPath, presetPath, hostLabel, presetLabel string) DiffResult {
	var hostContent string
	if data, err := os.ReadFile(hostPath); err == nil {
		hostContent = string(data)
	} else {
		hostContent = "(file does not exist on host)"
	}

	presetData, err := os.ReadFile(presetPath)
	if err != nil {
		return DiffResult{}
	}

	return GenerateUnifiedDiff(hostContent, string(presetData), hostLabel, presetLabel)
}
