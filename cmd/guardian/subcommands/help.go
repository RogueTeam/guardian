package subcommands

import (
	"flag"

	"github.com/RogueTeam/guardian/internal/commands"
)

var HelpCommand = commands.Command{
	Name:        "help",
	Aliases:     []string{"h", "-h", "--help"},
	Description: "Prints this help message",
	Callback: func(_ any, args []string) (result any, err error) {
		return nil, flag.ErrHelp
	},
}
