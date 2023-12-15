package subcommands

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
	"golang.org/x/term"
)

type SecretsConfig struct {
	Database         *string
	Get, Del, Set    *string
	ArgonTime        *uint
	ArgonThreads     *uint
	ArgonMemory      *uint
	NoPasswordPrompt *bool
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
		config.Get = set.String("get", "", "Search for the entry pointed by 'id'")
		config.Del = set.String("del", "", "Deletes the entry pointed by 'id'")
		config.Set = set.String("set", "", "Sets a new entry by its 'id'")
		// Argon settings
		config.ArgonTime = set.Uint("argon-time", uint(argon.Time), "Argon time to use")
		config.ArgonMemory = set.Uint("argon-memory", uint(argon.Memory), "Argon memory to use")
		config.ArgonThreads = set.Uint("argon-threads", uint(argon.Threads), "Argon threads to use")
		// Prompt
		config.NoPasswordPrompt = set.Bool("no-prompt", false, "No password prompt")
		return config
	},
	Callback: func(config any, args []string) (result any, err error) {
		dbConfig := config.(SecretsConfig)

		password := func() (key crypto.Data) {
			if !*dbConfig.NoPasswordPrompt {
				fmt.Fprint(os.Stderr, "Master key: ")
				defer fmt.Fprintln(os.Stderr, "")
			}
			password, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				log.Fatal(fmt.Errorf("failed to read password"))
			}
			copy(key.Buffer[:], password)
			key.Length = uint8(len(password))
			return key
		}

		file, err := os.OpenFile(*dbConfig.Database, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			err = fmt.Errorf("failed to open database: %s: %w", *dbConfig.Database, err)
			return
		}
		defer file.Close()

		db, err := database.Open(file)
		if err != nil {
			err = fmt.Errorf("failed to initialize database: %w", err)
			return
		}

		switch {
		case *dbConfig.Get != "":
			key := password()
			defer key.Release()

			var data crypto.Data
			data, err = db.GetSecret(*dbConfig.Get, &key)
			if err != nil {
				err = fmt.Errorf("failed to decrypt secret: %w", err)
				return
			}

			result = string(data.Bytes())
		case *dbConfig.Del != "":
			db.DelSecret(*dbConfig.Del)
			err = db.Save()
		case *dbConfig.Set != "":
			if len(args) == 0 {
				err = fmt.Errorf("failed to set secret, no secret provided in ARGS")
				return
			}
			var data crypto.Data
			defer data.Release()
			copy(data.Buffer[:], []byte(args[0]))
			data.Length = uint8(len([]byte(args[0])))

			argon := crypto.Argon{
				Time:    uint32(*dbConfig.ArgonTime),
				Memory:  uint32(*dbConfig.ArgonMemory),
				Threads: uint8(*dbConfig.ArgonThreads),
			}
			defer argon.Release()

			key := password()
			defer key.Release()

			err = db.SetSecret(*dbConfig.Set, &key, &data, &argon)
			if err == nil {
				err = db.Save()
			}
		}

		return
	},
}
