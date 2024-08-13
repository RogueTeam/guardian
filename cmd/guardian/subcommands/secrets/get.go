package secrets

import (
	"fmt"

	cliflags "github.com/RogueTeam/guardian/cmd/guardian/flags"
	"github.com/RogueTeam/guardian/cmd/guardian/utils"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var GetCommand = &commands.Command{
	Name:        "get",
	Description: "Retrieves a value from the database by its id",
	Args: commands.Values{
		{Type: commands.TypeString, Name: cliflags.Id, Description: "id of the entry"},
	},
	Setup: utils.SetupDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Dependencies
		db := ctx.MustGet(cliflags.Db).(*database.Database)

		// Retrieve
		result, err = db.Get(args[cliflags.Id].(string))
		if err != nil {
			err = fmt.Errorf("failed to retrieve value")
		}
		return
	},
}
