package secrets

import (
	"fmt"
	"os"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
	"github.com/RogueTeam/guardian/internal/utils/cli"
)

func openDBFile(ctx *commands.Context, flags map[string]any) (err error) {
	filepath := ctx.MustGet("secrets").(string)

	// Open file
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		err = fmt.Errorf("failed to open file: %s: %w", filepath, err)
		return
	}
	ctx.Set("file", file)

	// Setup argon
	argon := crypto.Argon{
		Time:    uint32(ctx.MustGet(ArgonTime).(int)),
		Memory:  uint32(ctx.MustGet(ArgonMemory).(int)),
		Threads: uint8(ctx.MustGet(ArgonThreads).(int)),
	}
	ctx.Set("argon", argon)

	// User key
	key := cli.ReadKey(!ctx.MustGet(NoPrompt).(bool))
	ctx.Set("key", key)

	return
}

func setupDB(ctx *commands.Context, flags map[string]any) (err error) {
	// Open DB file
	err = openDBFile(ctx, flags)
	if err != nil {
		err = fmt.Errorf("failed to open db file: %w", err)
		return
	}

	// Dependencies
	file := ctx.MustGet("file").(*os.File)
	key := ctx.MustGet("key").([]byte)

	// Prepare database
	db, err := database.Open(key, file)
	if err != nil {
		err = fmt.Errorf("failed to open database: %w", err)
		return
	}
	ctx.Set(Db, db)

	return
}

func deferSaveDB(ctx *commands.Context, result any) (finalResult any, err error) {
	finalResult = result

	// Dependencies
	file := ctx.MustGet("file").(*os.File)
	saltSize := ctx.MustGet("salt-size").(int)
	argon := ctx.MustGet("argon").(crypto.Argon)
	key := ctx.MustGet("key").([]byte)

	db := ctx.MustGet(Db).(*database.Database)

	// Save changes
	_, err = file.Seek(0, 0)
	if err == nil {
		err = db.Save(key, saltSize, &argon, file)
	}
	return
}
