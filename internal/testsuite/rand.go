package testsuite

import "crypto/rand"

func Random(length int) (buf []byte) {
	buf = make([]byte, length)
	rand.Read(buf)
	return buf
}
