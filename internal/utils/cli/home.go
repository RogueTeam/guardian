package cli

import (
	"log"
	"os/user"
)

func Home() string {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return u.HomeDir
}
