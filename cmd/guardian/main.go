package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/RogueTeam/guardian/internal/commands"
)

func init() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, commands.Usage(registered, os.Args[0]))
		os.Exit(1)
	}
}

func main() {
	result, err := commands.Handle(registered, os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr, commands.Usage(registered, os.Args[0]))
			return
		}
		slog.Error("Failed to execute command: %v", err)
		os.Exit(1)
	}

	switch result := result.(type) {
	case string, int, float64:
		fmt.Fprintf(os.Stdout, "%v", result)
	default:
		json.NewEncoder(os.Stdout).Encode(result)
	}
}
