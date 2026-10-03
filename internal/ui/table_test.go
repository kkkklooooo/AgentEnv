package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestTableRender(t *testing.T) {
	DisableColors()

	table := NewTable("NAME", "STATUS")
	table.AddRow("claude", "OK")
	table.AddRow("codex", "MISSING")

	var buf bytes.Buffer
	table.Render(&buf)

	output := buf.String()
	if !strings.Contains(output, "NAME") || !strings.Contains(output, "claude") {
		t.Fatalf("unexpected table output: %s", output)
	}
}
