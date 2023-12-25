package utils

import "log"

func Must[T any](a T, err error) (v T) {
	if err != nil {
		log.Fatal(err)
	}
	return a
}
