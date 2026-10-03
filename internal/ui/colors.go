package ui

import (
	"fmt"
	"os"
)

var noColor = false

func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		noColor = true
	}
}

// DisableColors forces ANSI colors off
func DisableColors() {
	noColor = true
}

func colorize(code int, s string) string {
	if noColor {
		return s
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, s)
}

func Green(s string) string  { return colorize(32, s) }
func Red(s string) string    { return colorize(31, s) }
func Yellow(s string) string { return colorize(33, s) }
func Blue(s string) string   { return colorize(34, s) }
func Magenta(s string) string{ return colorize(35, s) }
func Cyan(s string) string   { return colorize(36, s) }
func Gray(s string) string   { return colorize(90, s) }
func Bold(s string) string   { return colorize(1, s) }
