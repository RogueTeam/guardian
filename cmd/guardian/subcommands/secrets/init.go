package secrets

import (
	"fmt"
	"strings"

	cliflags "github.com/RogueTeam/guardian/cmd/guardian/flags"
	"github.com/RogueTeam/guardian/cmd/guardian/utils"
	"github.com/RogueTeam/guardian/crypto"
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

		config := database.Config{
			Key:      ctx.MustGet(cliflags.Key).([]byte),
			Argon:    ctx.MustGet(cliflags.Argon).(crypto.Argon),
			SaltSize: ctx.MustGet(cliflags.SaltSize).(int),
		}
		db, err := database.Open(config, strings.NewReader("{}"))
		if err != nil {
			err = fmt.Errorf("failed to initialize database: %w", err)
			return
		}
		ctx.Set(cliflags.Db, db)

		return
	},
}
