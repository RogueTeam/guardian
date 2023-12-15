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

var validCharactersBuffer = []byte(validCharacters)

// Only returns printable ascii characters from the random source
func PrintableRandom(b []byte) {
	max := big.NewInt(int64(len(validCharacters)))
	for index := range b {
		srcIndex, err := rand.Int(rand.Reader, max)
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
