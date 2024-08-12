package secrets

import (
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var SetCommand = &commands.Command{
	Name:        "set",
	Description: "Creates/Updates a entry",
	Args: commands.Values{
		{Type: commands.TypeString, Name: Id, Description: "id of the entry"},
		{Type: commands.TypeString, Name: Value, Description: "value of the entry"},
	},
	Setup: setupDB,
	Defer: deferSaveDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Dependencies
		db := ctx.MustGet(Db).(*database.Database)

		// Delete
		db.Set(args[Id].(string), args[Value].(string))
		return
	},
}
