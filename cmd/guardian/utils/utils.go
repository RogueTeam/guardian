package utils

import (
	"fmt"
	"os"

	cliflags "github.com/RogueTeam/guardian/cmd/guardian/flags"
	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
	"github.com/RogueTeam/guardian/internal/utils/cli"
)

func OpenDBFile(ctx *commands.Context, flags map[string]any) (err error) {
	filepath := ctx.MustGet(cliflags.Secrets).(string)

	// Open file
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		err = fmt.Errorf("failed to open file: %s: %w", filepath, err)
		return
	}
	ctx.Set(cliflags.File, file)

	// Setup argon
	argon := crypto.Argon{
		Time:    uint32(ctx.MustGet(cliflags.ArgonTime).(int)),
		Memory:  uint32(ctx.MustGet(cliflags.ArgonMemory).(int)),
		Threads: uint8(ctx.MustGet(cliflags.ArgonThreads).(int)),
	}
	ctx.Set(cliflags.Argon, argon)

	// User key
	key := cli.ReadKey(!ctx.MustGet(cliflags.NoPrompt).(bool))
	ctx.Set(cliflags.Key, key)

	return
}

func SetupDB(ctx *commands.Context, flags map[string]any) (err error) {
	// Open DB file
	err = OpenDBFile(ctx, flags)
	if err != nil {
		err = fmt.Errorf("failed to open db file: %w", err)
		return
	}

	// Dependencies
	file := ctx.MustGet(cliflags.File).(*os.File)
	key := ctx.MustGet(cliflags.Key).([]byte)

	// Prepare database
	config := database.Config{
		Key:      key,
		Argon:    ctx.MustGet(cliflags.Argon).(crypto.Argon),
		SaltSize: ctx.MustGet(cliflags.SaltSize).(int),
	}
	db, err := database.Open(config, file)
	if err != nil {
		err = fmt.Errorf("failed to open database: %w", err)
		return
	}
	ctx.Set(cliflags.Db, db)

	return
}

func DeferSaveDB(ctx *commands.Context, result any) (finalResult any, err error) {
	finalResult = result

	// Dependencies
	file := ctx.MustGet(cliflags.File).(*os.File)

	db := ctx.MustGet(cliflags.Db).(*database.Database)

	// Save changes
	_, err = file.Seek(0, 0)
	if err == nil {
		err = db.Save(file)
	}
	return
}
