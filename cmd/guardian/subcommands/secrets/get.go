package secrets

import (
	"fmt"

	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var GetCommand = &commands.Command{
	Name:        "get",
	Description: "Retrieves a value from the database by its id",
	Args: commands.Values{
		{Type: commands.TypeString, Name: Id, Description: "id of the entry"},
	},
	Setup: setupDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Dependencies
		db := ctx.MustGet(Db).(*database.Database)

		// Retrieve
		result, err = db.Get(args[Id].(string))
		if err != nil {
			err = fmt.Errorf("failed to retrieve value")
		}
		return
	},
}
