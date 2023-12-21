package subcommands

import (
	"fmt"
	"os"
	"path"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

var defaultArgon = crypto.DefaultArgon()

type SecretsConfig struct {
	Database             *string
	Get, Del, Set, Value *string
	SaltSize             *int
	ArgonTime            *uint
	ArgonThreads         *uint
	ArgonMemory          *uint
	NoPasswordPrompt     *bool
	Init                 *bool
}

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
		Time:    uint32(ctx.MustGet("argon-time").(int)),
		Memory:  uint32(ctx.MustGet("argon-memory").(int)),
		Threads: uint8(ctx.MustGet("argon-threads").(int)),
	}
	ctx.Set("argon", argon)

	// User key
	key := ReadKey(!ctx.MustGet("no-prompt").(bool))
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
	ctx.Set("db", db)

	return
}

func deferSaveDB(ctx *commands.Context, result any) (finalResult any, err error) {
	finalResult = result

	// Dependencies
	file := ctx.MustGet("file").(*os.File)
	saltSize := ctx.MustGet("salt-size").(int)
	argon := ctx.MustGet("argon").(crypto.Argon)
	key := ctx.MustGet("key").([]byte)

	db := ctx.MustGet("db").(*database.Database)

	// Save changes
	_, err = file.Seek(0, 0)
	if err == nil {
		err = db.Save(key, saltSize, &argon, file)
	}
	return
}

var SecretsCommand = &commands.Command{
	Name:        "secrets",
	Description: "Manipulate the database JSON file",
	Flags: commands.Values{
		{Type: commands.TypeString, Name: "secrets", Description: "Secrets database to use", Default: path.Join(home(), "guardian.json")},
		{Type: commands.TypeInt, Name: "salt-size", Description: "Size of the random salt to read", Default: 1024},
		{Type: commands.TypeInt, Name: "argon-time", Description: "Argon time config", Default: int(defaultArgon.Time)},
		{Type: commands.TypeInt, Name: "argon-memory", Description: "Argon memory config", Default: int(defaultArgon.Memory)},
		{Type: commands.TypeInt, Name: "argon-threads", Description: "Argon threads config", Default: int(defaultArgon.Threads)},
		{Type: commands.TypeBool, Name: "no-prompt", Description: "No password prompt", Default: false},
	},
	Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
		ctx.Set("secrets", flags["secrets"])
		ctx.Set("salt-size", flags["salt-size"])
		ctx.Set("argon-time", flags["argon-time"])
		ctx.Set("argon-memory", flags["argon-memory"])
		ctx.Set("argon-threads", flags["argon-threads"])
		ctx.Set("no-prompt", flags["no-prompt"])

		return
	},
	SubCommands: commands.Commands{
		{
			Name:        "init",
			Description: "Initialize the secrets database",
			Setup:       openDBFile,
			Defer:       deferSaveDB,
			Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
				// Initialize command
				var db = database.New()
				ctx.Set("db", db)

				return
			},
		},
		{
			Name:        "get",
			Description: "Retrieves a value from the database by its id",
			Args: commands.Values{
				{Type: commands.TypeString, Name: "id", Description: "id of the entry"},
			},
			Setup: setupDB,
			Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
				// Dependencies
				db := ctx.MustGet("db").(*database.Database)

				// Retrieve
				result, err = db.Get(args["id"].(string))
				if err != nil {
					err = fmt.Errorf("failed to retrieve value")
				}
				return
			},
		},
		{
			Name:        "del",
			Description: "Deletes a entry from the database by its id",
			Args: commands.Values{
				{Type: commands.TypeString, Name: "id", Description: "id of the entry"},
			},
			Setup: setupDB,
			Defer: deferSaveDB,
			Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
				// Dependencies
				db := ctx.MustGet("db").(*database.Database)

				// Delete
				err = db.Del(args["id"].(string))
				if err != nil {
					err = fmt.Errorf("failed to delete value")
				}
				return
			},
		},
		{
			Name:        "set",
			Description: "Creates/Updates a entry",
			Args: commands.Values{
				{Type: commands.TypeString, Name: "id", Description: "id of the entry"},
				{Type: commands.TypeString, Name: "value", Description: "value of the entry"},
			},
			Setup: setupDB,
			Defer: deferSaveDB,
			Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
				// Dependencies
				db := ctx.MustGet("db").(*database.Database)

				// Delete
				db.Set(args["id"].(string), args["value"].(string))
				return
			},
		},
	},
}
