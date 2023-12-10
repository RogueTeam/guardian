package crypto

import "crypto/rand"

func DefaultArgon() (a Argon) {
	rand.Read(a.Salt[:])
	a.Memory = 64 * 1024
	a.Time = 1024
	a.Threads = 64

	return a
}
