package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/sha3"
)

const (
	ChecksumSize = 512 / 8
	ChunkSize    = 256
	DataSize     = ChunkSize - 1
	KeySize      = 32
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

func (a *Argon) Release() {
	rand.Read(a.Salt[:])
	a.Time = 0
	a.Memory = 0
	a.Threads = 0
}

// Stretches the unique password for the job based on the master password
func (a Argon) Stretch(k []byte) (stretch []byte) {
	stretch = argon2.IDKey(k, a.Salt[:], a.Time, a.Memory, a.Threads, KeySize)
	return stretch
}

// Struct to manage the data in the data, removing any possibility of heap allocation
type Data struct {
	Buffer [DataSize]byte
	Length uint8
}

func (d *Data) Bytes() []byte {
	return d.Buffer[:d.Length]
}

func (d *Data) Release() {
	rand.Read(d.Buffer[:])
	d.Length = 0
}

func RandomData() (d Data) {
	PrintableRandom(d.Buffer[:])
	d.Length = DataSize
	return d
}

type Secret struct {
	IV            [aes.BlockSize]byte
	Cipher        [ChunkSize]byte
	KeyArgon      Argon
	Checksum      [ChecksumSize]byte
	ChecksumArgon Argon
}

func (s *Secret) Release() {
	rand.Read(s.IV[:])
	rand.Read(s.Cipher[:])
	rand.Read(s.Checksum[:])

	s.KeyArgon.Release()
	s.ChecksumArgon.Release()
}

var ErrNotPrintable = errors.New("non printable characters")

// Encrypts the received password and the returns all the information of the encrypted block
func Encrypt(key, data Data, argon Argon) (secret Secret, err error) {
	defer key.Release()
	defer argon.Release()

	for _, value := range key.Bytes() {
		if !isPrintable(value) {
			err = fmt.Errorf("%w: found non printable in key", ErrNotPrintable)
			return
		}
	}

	for _, value := range data.Bytes() {
		if !isPrintable(value) {
			err = fmt.Errorf("%w: found non printable in secret", ErrNotPrintable)
			return
		}
	}

	// Stretched key
	// Setup argon
	secret.KeyArgon = argon
	stretchKey := secret.KeyArgon.Stretch(key.Bytes())
	rand.Read(secret.IV[:])

	// Prepare the buffer to be encrypted
	// Protect memory after been used
	// Do not wait for the GC to collect it
	buffer := make([]byte, ChunkSize)
	defer rand.Read(buffer)

	buffer[0] = data.Length
	n := copy(buffer[1:], data.Bytes())
	PrintableRandom(buffer[1+n:])

	// Checksum Buffer
	// Use the same settings as the received Argon settings
	// This way we ensure the checksum is also computational possible
	secret.ChecksumArgon = secret.KeyArgon
	rand.Read(secret.ChecksumArgon.Salt[:])
	stretchBuffer := secret.ChecksumArgon.Stretch(buffer)

	// Checksum buffer
	hash := sha3.New512()
	hash.Write(stretchBuffer)
	copy(secret.Checksum[:], hash.Sum(nil))

	// This is impossible to panic since stretch is ensured to be of the valid fixed KeySize
	block, _ := aes.NewCipher(stretchKey)
	cbc := cipher.NewCBCEncrypter(block, secret.IV[:])
	cbc.CryptBlocks(secret.Cipher[:], buffer)

	return
}

var ErrInvalidPassword = errors.New("invalid password")

// Decrypt the received secret
func Decrypt(key Data, s Secret) (data Data, err error) {
	defer key.Release()
	defer s.Release()

	stretchKey := s.KeyArgon.Stretch(key.Bytes())

	buffer := make([]byte, ChunkSize)
	defer rand.Read(buffer)

	block, _ := aes.NewCipher(stretchKey)
	cbc := cipher.NewCBCDecrypter(block, s.IV[:])
	cbc.CryptBlocks(buffer, s.Cipher[:])

	// Stretch buffer
	stretchBuffer := s.ChecksumArgon.Stretch(buffer)

	// Calculate checksum
	computedChecksum := make([]byte, ChecksumSize)
	defer rand.Read(computedChecksum)

	hash := sha3.New512()
	hash.Write(stretchBuffer)
	copy(computedChecksum, hash.Sum(nil))

	if !bytes.Equal(computedChecksum, s.Checksum[:]) {
		err = ErrInvalidPassword
		return
	}

	copy(data.Buffer[:], buffer[1:])
	data.Length = buffer[0]

	return
}
