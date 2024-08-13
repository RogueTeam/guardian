package main

import (
	"github.com/RogueTeam/guardian/cmd/guardian/subcommands/mount"
	"github.com/RogueTeam/guardian/cmd/guardian/subcommands/secrets"
	"github.com/RogueTeam/guardian/internal/commands"
)

var root = commands.Command{
	Name:        "guardian",
	Description: "Your portable personal file guardian",
	SubCommands: commands.Commands{
		secrets.SecretsCommand,
		mount.MountCommand,
	},
}
