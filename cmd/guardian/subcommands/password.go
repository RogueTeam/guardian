package subcommands

import (
	"encoding/base64"
	"encoding/hex"
	"flag"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/internal/commands"
)

type PasswordConfig struct {
	Length *int
	Base64 *bool
	Hex    *bool
}

var PasswordCommand = commands.Command{
	Name:        "password",
	Aliases:     []string{"pass"},
	Description: "Password generator",
	Init: func(set *flag.FlagSet) any {
		var passwordConfig PasswordConfig
		passwordConfig.Length = set.Int("length", 16, "Length of the password to generate")
		passwordConfig.Base64 = set.Bool("base64", false, "Encode the final password to base64")
		passwordConfig.Hex = set.Bool("hex", false, "Encode the final password to hex")
		return passwordConfig
	},
	Callback: func(config any, args []string) (result any, err error) {
		pConfig := config.(PasswordConfig)

		password := make([]byte, *pConfig.Length)
		crypto.PrintableRandom(password)

		switch {
		case *pConfig.Base64:
			result = base64.StdEncoding.EncodeToString(password)
		case *pConfig.Hex:
			result = hex.EncodeToString(password)
		default:
			result = string(password)
		}
		return result, err
	},
}
