package secrets

import (
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var InitCommand = &commands.Command{
	Name:        "init",
	Description: "Initialize the secrets database",
	Setup:       openDBFile,
	Defer:       deferSaveDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Initialize command
		var db = database.New()
		ctx.Set(Db, db)

		return
	},
}
