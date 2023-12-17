package subcommands

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
)

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

var SecretsCommand = commands.Command{
	Name:        "secrets",
	Description: "Manipulate the database JSON file",
	Init: func(set *flag.FlagSet) any {
		argon := crypto.DefaultArgon()

		curr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		var config SecretsConfig
		config.Database = set.String("database", path.Join(curr.HomeDir, "guardian.json"), "Database file to use")

		// Get existing entry
		config.Get = set.String("get", "", "Search for the entry pointed by 'id'")

		// Delete existing entry
		config.Del = set.String("del", "", "Deletes the entry pointed by 'id'")

		// Set new entry
		config.Set = set.String("set", "", "Sets a new entry by its 'id'")
		config.Value = set.String("value", "", "Value to use during -set")
		// Salt configuration
		config.SaltSize = set.Int("salt-size", 1024, "Salt size to use during save")

		// Argon settings
		config.ArgonTime = set.Uint("argon-time", uint(argon.Time), "Argon time to use")
		config.ArgonMemory = set.Uint("argon-memory", uint(argon.Memory), "Argon memory to use")
		config.ArgonThreads = set.Uint("argon-threads", uint(argon.Threads), "Argon threads to use")

		// Prompt
		config.NoPasswordPrompt = set.Bool("no-prompt", false, "No password prompt")

		// Initialize database
		config.Init = set.Bool("init", false, "Initialize a database file")
		return config
	},
	Callback: func(config any, args []string) (result any, err error) {
		// Cast received configuration
		dbConfig := config.(SecretsConfig)

		// Open file
		file, err := os.OpenFile(*dbConfig.Database, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			err = fmt.Errorf("failed to open file: %s: %w", *dbConfig.Database, err)
			return
		}
		defer file.Close()

		// User key
		key := ReadKey(!*dbConfig.NoPasswordPrompt)
		defer rand.Read(key)

		// Setup argon
		argon := crypto.Argon{
			Time:    uint32(*dbConfig.ArgonTime),
			Memory:  uint32(*dbConfig.ArgonMemory),
			Threads: uint8(*dbConfig.ArgonThreads),
		}

		// Initialize command
		if *dbConfig.Init {
			var db = database.New()
			err = db.Save(key, *dbConfig.SaltSize, &argon, file)
			return
		}

		// Open Database
		db, err := database.Open(key, file)
		if err != nil {
			err = fmt.Errorf("failed to open database: %w", err)
			return
		}

		switch {
		case *dbConfig.Get != "":
			result, err = db.Get(*dbConfig.Get)
		case *dbConfig.Del != "":
			err = db.Del(*dbConfig.Del)
			if err != nil {
				return
			}

			_, err = file.Seek(0, 0)
			if err == nil {
				err = db.Save(key, *dbConfig.SaltSize, &argon, file)
			}
		case *dbConfig.Set != "":
			db.Set(*dbConfig.Set, *dbConfig.Value)

			_, err = file.Seek(0, 0)
			if err == nil {
				err = db.Save(key, *dbConfig.SaltSize, &argon, file)
			}
		}

		return
	},
}
