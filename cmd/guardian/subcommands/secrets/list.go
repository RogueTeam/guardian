package secrets

import (
	"fmt"

	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var ListCommand = &commands.Command{
	Name:        "list",
	Description: "List all available keys",
	Args: commands.Values{
		{Type: commands.TypeString, Name: Id, Description: "id of the entry"},
	},
	Setup: setupDB,
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		// Dependencies
		db := ctx.MustGet(Db).(*database.Database)

		// Retrieve
		names, err := db.List()
		if err != nil {
			err = fmt.Errorf("failed to list entries: %w", err)
			return
		}

		result = names
		return
	},
}
