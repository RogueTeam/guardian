// Subcommand for generating random passwords
package password

import "github.com/RogueTeam/guardian/crypto"

func New(length uint8) (password []byte) {
	password = make([]byte, length)

	crypto.PrintableRandom(password)

	return password
}
