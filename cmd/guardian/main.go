package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	result, err := root.Run(os.Args[1:])
	if err != nil {
		log.Fatalf("something went wrong: %v", err)
	}

	switch result := result.(type) {
	case nil:
		fmt.Fprintf(os.Stderr, "Command exited successfully!")
	case string, int, float64:
		fmt.Fprintf(os.Stdout, "%v", result)
	default:
		json.NewEncoder(os.Stdout).Encode(result)
	}
}
