package secrets

import (
	cliflags "github.com/RogueTeam/guardian/cmd/guardian/flags"
	"github.com/RogueTeam/guardian/cmd/guardian/utils"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var SetCommand = &commands.Command{
	Name:        "set",
	Description: "Creates/Updates a entry",
	Args: commands.Values{
		{Type: commands.TypeString, Name: cliflags.Id, Description: "id of the entry"},
		{Type: commands.TypeString, Name: cliflags.Value, Description: "value of the entry"},
	},
	Setup: utils.SetupDB,
	Defer: utils.DeferSaveDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Dependencies
		db := ctx.MustGet(cliflags.Db).(*database.Database)

		// Delete
		db.Set(args[cliflags.Id].(string), args[cliflags.Value].(string))
		return
	},
}
