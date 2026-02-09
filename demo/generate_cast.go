//go:build ignore

// generate_cast.go generates an asciicast v2 (.cast) file for the sekret demo.
//
// Usage:
//
//	go run demo/generate_cast.go > demo/demo.cast
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	cols = 80
	rows = 15

	charDelay     = 0.04  // 40ms per character (uniform for all typing)
	backspaceRate = 0.005 // 5ms per backspace
	prompt        = "$ "
)

type header struct {
	Version int               `json:"version"`
	Width   int               `json:"width"`
	Height  int               `json:"height"`
	Env     map[string]string `json:"env"`
}

// cursor tracks the current time in seconds.
var cursor float64

func emit(delay float64, text string) {
	cursor += delay
	b, _ := json.Marshal([]interface{}{round3(cursor), "o", text})
	fmt.Fprintln(os.Stdout, string(b))
}

func round3(f float64) float64 {
	return float64(int(f*1000)) / 1000
}

// typeText simulates typing each character with uniform charDelay.
func typeText(text string) {
	for _, ch := range text {
		emit(charDelay, string(ch))
	}
}

// outputLine prints program output instantly with newline.
func outputLine(text string) {
	emit(0, text+"\r\n")
}

// outputText prints program output instantly without newline.
func outputText(text string) {
	emit(0, text)
}

// pressEnter simulates pressing enter.
func pressEnter() {
	emit(0, "\r\n")
}

// backspace simulates n backspaces with visual deletion.
func backspace(n int) {
	for i := 0; i < n; i++ {
		emit(backspaceRate, "\b \b")
	}
}

// sleep adds a pause.
func sleep(seconds float64) {
	cursor += seconds
}

// showPrompt outputs the shell prompt.
func showPrompt() {
	emit(0, prompt)
}

// clearScreen sends ANSI clear.
func clearScreen() {
	emit(0, "\x1b[2J\x1b[H")
}

// round describes one iteration of the demo (export → sekret add → list → env).
type round struct {
	shorthand  string // e.g. "anthropic"
	envVar     string // e.g. "ANTHROPIC_API_KEY"
	exportLine string // e.g. `export ANTHROPIC_API_KEY="sk-ant-`
	maskLen    int    // number of '*' characters for masked key input
	listFile   string // path to captured sekret list output, e.g. "demo/list_1.txt"
}

// loadListLines reads the captured sekret list output file and returns non-empty lines.
func loadListLines(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot open %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func main() {
	// Header
	h := header{
		Version: 2,
		Width:   cols,
		Height:  rows,
		Env: map[string]string{
			"SHELL": "/bin/zsh",
			"TERM":  "xterm-256color",
		},
	}
	hb, _ := json.Marshal(h)
	fmt.Fprintln(os.Stdout, string(hb))

	// List output is read from files captured by `go run demo/capture_list.go`.
	rounds := []round{
		{
			shorthand:  "anthropic",
			envVar:     "ANTHROPIC_API_KEY",
			exportLine: `export ANTHROPIC_API_KEY="sk-ant-`,
			maskLen:    16,
			listFile:   "demo/list_1.txt",
		},
		{
			shorthand:  "gemini",
			envVar:     "GEMINI_API_KEY",
			exportLine: `export GEMINI_API_KEY="AIza`,
			maskLen:    16,
			listFile:   "demo/list_2.txt",
		},
		{
			shorthand:  "openai",
			envVar:     "OPENAI_API_KEY",
			exportLine: `export OPENAI_API_KEY="sk-proj-`,
			maskLen:    16,
			listFile:   "demo/list_3.txt",
		},
	}

	for i, r := range rounds {
		isLast := i == len(rounds)-1

		// --- Type the export line (user hesitates about plaintext) ---
		showPrompt()
		sleep(0.5)
		typeText(r.exportLine)
		sleep(0.8)

		// --- Backspace to erase it ---
		lineLen := len(r.exportLine)
		backspace(lineLen)
		sleep(0.3)

		// --- Type sekret add <shorthand> ---
		typeText("sekret add " + r.shorthand)
		pressEnter()

		// --- Shorthand confirmation prompt ---
		outputText(fmt.Sprintf(`  "%s" -> %s. Continue? [Y/n]: `, r.shorthand, r.envVar))
		sleep(0.3)
		typeText("y")
		pressEnter()

		// --- Password prompt ---
		outputText("  API Key: ")
		sleep(0.3)

		// Masked input (asterisks)
		for range r.maskLen {
			emit(charDelay, "*")
		}
		pressEnter()

		// --- Saved message ---
		outputLine(fmt.Sprintf("  Saved to OS keychain (%s)", r.envVar))

		// --- sekret list ---
		showPrompt()
		sleep(0.8)
		typeText("sekret list")
		pressEnter()

		listLines := loadListLines(r.listFile)
		for _, line := range listLines {
			outputLine(line)
		}

		// --- eval "$(sekret env)" ---
		showPrompt()
		sleep(1.0)
		typeText(`eval "$(sekret env)"`)
		pressEnter()

		// --- # loaded message ---
		showPrompt()
		sleep(0.3)
		typeText(fmt.Sprintf("# $%s loaded!", r.envVar))
		sleep(1.5)

		if !isLast {
			// Clear screen for next round
			sleep(0.5)
			clearScreen()
		}
	}

	// Final pause so the last frame stays visible before looping.
	// Emit a no-op event after sleeping so svg-term includes the pause in its duration.
	sleep(3.0)
	emit(0, "")
}
