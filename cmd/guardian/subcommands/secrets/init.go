package secrets

import (
	cliflags "github.com/RogueTeam/guardian/cmd/guardian/flags"
	"github.com/RogueTeam/guardian/cmd/guardian/utils"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var InitCommand = &commands.Command{
	Name:        "init",
	Description: "Initialize the secrets database",
	Setup:       utils.OpenDBFile,
	Defer:       utils.DeferSaveDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Initialize command
		var db = database.New()
		ctx.Set(cliflags.Db, db)

		return
	},
}
