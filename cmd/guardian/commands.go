package main

import (
	"github.com/RogueTeam/guardian/cmd/guardian/subcommands"
	"github.com/RogueTeam/guardian/internal/commands"
)

var registered = []commands.Command{
	subcommands.HelpCommand,
	subcommands.PasswordCommand,
}
