package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noma4i/westpac-cli/api"
	"github.com/noma4i/westpac-cli/utils"
	"github.com/noma4i/westpac-cli/views"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging to ~/.westpac/debug.log")
	mask := flag.Bool("mask", false, "Mask account names and descriptions (keep amounts visible)")
	flag.Parse()

	utils.MaskMode = *mask

	var debugLog func(string)
	if *debug {
		logDir, _ := api.ConfigDir()
		os.MkdirAll(logDir, 0700)
		logPath := filepath.Join(logDir, "debug.log")

		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatalf("Cannot open log file %s: %v", logPath, err)
		}
		defer f.Close()

		fmt.Fprintf(f, "\n=== Session started %s ===\n", time.Now().Format(time.RFC3339))
		debugLog = func(msg string) {
			fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05.000"), msg)
		}
		fmt.Fprintf(os.Stderr, "Debug log: %s\n", logPath)
	}

	client, err := api.NewClient("", *debug, debugLog)
	if err != nil {
		log.Fatal(err)
	}

	app := views.NewAppModel(client)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
