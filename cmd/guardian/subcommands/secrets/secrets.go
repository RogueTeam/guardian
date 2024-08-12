package secrets

import (
	"path"

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
	Name:        Secrets,
	Description: "Manipulate the database JSON file",
	Flags: commands.Values{
		{Type: commands.TypeString, Name: Secrets, Description: "Secrets database to use", Default: path.Join(cli.Home(), "guardian.json")},
		{Type: commands.TypeInt, Name: SaltSize, Description: "Size of the random salt to read", Default: crypto.DefaultSaltSize},
		{Type: commands.TypeInt, Name: ArgonTime, Description: "Argon time config", Default: int(defaultArgon.Time)},
		{Type: commands.TypeInt, Name: ArgonMemory, Description: "Argon memory config", Default: int(defaultArgon.Memory)},
		{Type: commands.TypeInt, Name: ArgonThreads, Description: "Argon threads config", Default: int(defaultArgon.Threads)},
		{Type: commands.TypeBool, Name: NoPrompt, Description: "No password prompt", Default: false},
	},
	Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
		ctx.Set(Secrets, flags[Secrets])
		ctx.Set(SaltSize, flags[SaltSize])
		ctx.Set(ArgonTime, flags[ArgonTime])
		ctx.Set(ArgonMemory, flags[ArgonMemory])
		ctx.Set(ArgonThreads, flags[ArgonThreads])
		ctx.Set(NoPrompt, flags[NoPrompt])

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
