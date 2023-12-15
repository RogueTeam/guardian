package crypto

func DefaultArgon() (a Argon) {
	a.Memory = 64 * 1024
	a.Time = 1024
	a.Threads = 64

	return a
}
