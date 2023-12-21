package subcommands

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"golang.org/x/term"
)

func home() string {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return u.HomeDir
}

func ReadKey(prompt bool) []byte {
	if prompt {
		fmt.Fprint(os.Stderr, "Master key: ")
		defer fmt.Fprintln(os.Stderr, "")
	}
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read password"))
	}
	return password
}
