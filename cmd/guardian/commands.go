package main

import (
	"github.com/RogueTeam/guardian/cmd/guardian/subcommands"
	"github.com/RogueTeam/guardian/internal/commands"
)

var root = commands.Command{
	Name:        "guardian",
	Description: "Your portable personal file guardian",
	SubCommands: commands.Commands{
		subcommands.SecretsCommand,
	},
}
