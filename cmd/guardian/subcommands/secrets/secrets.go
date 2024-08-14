package secrets

import (
	"path"

	cliflags "github.com/RogueTeam/guardian/cmd/guardian/flags"
	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/internal/commands"
	"github.com/RogueTeam/guardian/internal/utils/cli"
)

var defaultArgon = crypto.DefaultArgon()

type SecretsConfig struct {
	Database             *string
	Get, Del, Set, Value *string
	SaltSize             *int
	ArgonTime            *uint
	ArgonThreads         *uint
	ArgonMemory          *uint
	NoPasswordPrompt     *bool
	Init                 *bool
}

var SecretsCommand = &commands.Command{
	Name:        cliflags.Secrets,
	Description: "Manipulate the database JSON file",
	Flags: commands.Values{
		{Type: commands.TypeString, Name: cliflags.Secrets, Description: "Secrets database to use", Default: path.Join(cli.Home(), "guardian.json")},
		{Type: commands.TypeInt, Name: cliflags.SaltSize, Description: "Size of the random salt to read", Default: crypto.DefaultSaltSize},
		{Type: commands.TypeInt, Name: cliflags.ArgonTime, Description: "Argon time config", Default: int(defaultArgon.Time)},
		{Type: commands.TypeInt, Name: cliflags.ArgonMemory, Description: "Argon memory config", Default: int(defaultArgon.Memory)},
		{Type: commands.TypeInt, Name: cliflags.ArgonThreads, Description: "Argon threads config", Default: int(defaultArgon.Threads)},
		{Type: commands.TypeBool, Name: cliflags.NoPrompt, Description: "No password prompt", Default: false},
	},
	Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
		ctx.Set(cliflags.Secrets, flags[cliflags.Secrets])
		ctx.Set(cliflags.SaltSize, flags[cliflags.SaltSize])
		ctx.Set(cliflags.ArgonTime, flags[cliflags.ArgonTime])
		ctx.Set(cliflags.ArgonMemory, flags[cliflags.ArgonMemory])
		ctx.Set(cliflags.ArgonThreads, flags[cliflags.ArgonThreads])
		ctx.Set(cliflags.NoPrompt, flags[cliflags.NoPrompt])

		return
	},
	SubCommands: commands.Commands{
		InitCommand,
		GetCommand,
		ListCommand,
		DelCommand,
		SetCommand,
	},
}
