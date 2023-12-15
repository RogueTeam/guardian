package crypto

import (
	"crypto/rand"
	"log/slog"
	"math/big"
)

func isPrintable(c byte) bool {
	return c >= 33 && c <= 126
}

const (
	lower           = "abcdefghijklmnopqrstuvwxyz"
	upper           = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits          = "0123456789"
	symbols         = "!@#$%^&*()_+-=[]{}\\|:;'\",.<>/?`~"
	validCharacters = lower + upper + digits + symbols
)

var (
	validCharactersBuffer = []byte(validCharacters)
	countValidCharacters  = big.NewInt(int64(len(validCharacters)))
)

// Only returns printable ascii characters from the random source
func PrintableRandom(b []byte) {
	for index := range b {
		srcIndex, err := rand.Int(rand.Reader, countValidCharacters)
		if err != nil {
			slog.Error("Failed to read random data")
		}
		b[index] = validCharactersBuffer[srcIndex.Int64()]
	}
}

func RandomData() (d Data) {
	PrintableRandom(d.Buffer[:])
	d.Length = DataSize
	return d
}
