package secrets

import (
	"fmt"

	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var DelCommand = &commands.Command{
	Name:        "del",
	Description: "Deletes a entry from the database by its id",
	Args: commands.Values{
		{Type: commands.TypeString, Name: Id, Description: "id of the entry"},
	},
	Setup: setupDB,
	Defer: deferSaveDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Dependencies
		db := ctx.MustGet(Db).(*database.Database)

		// Delete
		err = db.Del(args[Id].(string))
		if err != nil {
			err = fmt.Errorf("failed to delete value")
		}
		return
	},
}
