package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/sha3"
)

const (
	ChunkSize  = 256
	SecretSize = ChunkSize - 1
	KeySize    = 32
)

func isPrintable(c byte) bool {
	return c >= 33 && c <= 126
}

// Only returns printable ascii characters from the random source
func PrintableRandom(b []byte) {
	var random [ChunkSize]byte
	defer rand.Read(random[:])

	for index := 0; index < len(b); {
		rand.Read(random[:])
		for _, value := range random {
			if index >= len(b) {
				break
			}
			if isPrintable(value) {
				b[index] = value
				index++
			}
		}
	}
}

type Argon struct {
	Salt    [ChunkSize]byte
	Time    uint32
	Memory  uint32
	Threads uint8
}

// Stretches the unique password for the job based on the master password
func (a Argon) Stretch(k []byte) (stretch []byte) {
	stretch = argon2.IDKey(k, a.Salt[:], a.Time, a.Memory, a.Threads, KeySize)
	return stretch
}

// Struct to manage the data in the data, removing any possibility of heap allocation
type Data struct {
	Buffer [ChunkSize]byte
	Length int
}

func (d *Data) Release() {
	rand.Read(d.Buffer[:])
	d.Length = 0
}

func RandomData() (d Data) {
	PrintableRandom(d.Buffer[:])
	d.Length = SecretSize
	return d
}

type Secret struct {
	IV           [aes.BlockSize]byte
	Cipher       [ChunkSize]byte
	ChecksumSalt [ChunkSize]byte
	Checksum     [512 / 8]byte
	Argon        Argon
}

func (e *Secret) Release() {
	rand.Read(e.IV[:])
	rand.Read(e.Cipher[:])
	rand.Read(e.ChecksumSalt[:])
	rand.Read(e.Checksum[:])
}

var (
	ErrInvalidSecretSize = errors.New("invalid secret size")
	ErrNotPrintable      = errors.New("non printable characters")
)

// Encrypts the received password and the returns all the information of the encrypted block
func Encrypt(key, secret Data, argon Argon) (s Secret, err error) {
	if secret.Length > SecretSize {
		err = fmt.Errorf("%w: max secret size is %d", ErrInvalidSecretSize, SecretSize)
		return
	} else if secret.Length == 0 {
		err = fmt.Errorf("%w: secret cannot be empty", ErrInvalidSecretSize)
		return
	}

	for _, value := range key.Buffer[:key.Length] {
		if !isPrintable(value) {
			err = fmt.Errorf("%w: found non printable in key", ErrNotPrintable)
			return
		}
	}

	for _, value := range secret.Buffer[:secret.Length] {
		if !isPrintable(value) {
			err = fmt.Errorf("%w: found non printable in secret", ErrNotPrintable)
			return
		}
	}

	// Stretched key
	stretch := argon.Stretch(key.Buffer[:key.Length])
	rand.Read(s.IV[:])

	// Prepare the buffer to be encryptoed
	b := make([]byte, ChunkSize)
	b[0] = uint8(secret.Length)
	n := copy(b[1:], secret.Buffer[:secret.Length])
	PrintableRandom(b[1+n:])

	// Protect memory after been used
	// Do not wait for the GC to collect it
	defer rand.Read(b)

	// Checksum Buffer
	// TODO: Update whitepaper using this new hash algorithm
	// TODO: https://crypto.stackexchange.com/questions/98969/will-hashing-multiple-times-be-more-less-or-similarly-secure-as-hashing-once
	rand.Read(s.ChecksumSalt[:])
	hash := sha3.New512()
	hash.Write(s.ChecksumSalt[:])
	hash.Write(s.Cipher[:])
	copy(s.Checksum[:], hash.Sum(nil))

	// Setup argon
	s.Argon = argon

	// This is impossible to panic since stretch is ensured to be of the valid fixed KeySize
	block, _ := aes.NewCipher(stretch)
	cbc := cipher.NewCBCEncrypter(block, s.IV[:])
	cbc.CryptBlocks(s.Cipher[:], b)

	return
}

func Decrypt(key, s Secret) {}
