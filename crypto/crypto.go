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

type Argon struct {
	Time    uint32 `json:"time"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
}

func (a *Argon) Release() {
	a.Time = 0
	a.Memory = 0
	a.Threads = 0
}

// Stretches the unique password for the job based on the master password
func (a Argon) Stretch(k []byte, salt []byte) (stretch []byte) {
	stretch = argon2.IDKey(k, salt, a.Time, a.Memory, a.Threads, KeySize)
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

type Secret struct {
	IV           [aes.BlockSize]byte `json:"iv"`
	Cipher       [ChunkSize]byte     `json:"cipher"`
	Checksum     [ChecksumSize]byte  `json:"checksum"`
	KeySalt      [ChunkSize]byte     `json:"keySalt"`
	ChecksumSalt [ChunkSize]byte     `json:"checksumSalt"`
	Argon        Argon               `json:"argon"`
}

func (s *Secret) Release() {
	rand.Read(s.IV[:])
	rand.Read(s.Cipher[:])
	rand.Read(s.Checksum[:])
	rand.Read(s.KeySalt[:])
	rand.Read(s.ChecksumSalt[:])

	s.Argon.Release()
}

// Init initialize IV and Salts
func (s *Secret) Init() {
	rand.Read(s.IV[:])
	rand.Read(s.KeySalt[:])
	rand.Read(s.ChecksumSalt[:])
}

var ErrNotPrintable = errors.New("non printable characters")

// Encrypts the received secret
func Encrypt(key, data *Data, argon *Argon) (secret Secret, err error) {
	defer func() {
		if err != nil {
			secret.Release()
		}
	}()

	// Verify secure key
	for _, value := range key.Bytes() {
		if !isPrintable(value) {
			err = fmt.Errorf("%w: found non printable in key", ErrNotPrintable)
			return
		}
	}

	// Verify secure data
	for _, value := range data.Bytes() {
		if !isPrintable(value) {
			err = fmt.Errorf("%w: found non printable in secret", ErrNotPrintable)
			return
		}
	}

	// Initialize secret
	secret.Init()

	// Setup argon
	secret.Argon = *argon

	// Stretched key
	stretchKey := secret.Argon.Stretch(key.Bytes(), secret.KeySalt[:])

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
	rand.Read(secret.ChecksumSalt[:])
	stretchBuffer := secret.Argon.Stretch(buffer, secret.ChecksumSalt[:])

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
func Decrypt(key *Data, secret *Secret) (data Data, err error) {
	defer func() {
		if err != nil {
			data.Release()
		}
	}()

	// Stretch key
	stretchKey := secret.Argon.Stretch(key.Bytes(), secret.KeySalt[:])

	// Reserve decryption buffer
	buffer := make([]byte, ChunkSize)
	defer rand.Read(buffer)

	// Decrypt
	block, _ := aes.NewCipher(stretchKey)
	cbc := cipher.NewCBCDecrypter(block, secret.IV[:])
	cbc.CryptBlocks(buffer, secret.Cipher[:])

	// Stretch buffer
	stretchBuffer := secret.Argon.Stretch(buffer, secret.ChecksumSalt[:])

	// Calculate checksum
	computedChecksum := make([]byte, ChecksumSize)
	defer rand.Read(computedChecksum)

	// Calculate checksum
	hash := sha3.New512()
	hash.Write(stretchBuffer)
	copy(computedChecksum, hash.Sum(nil))

	// Verify checksum
	if !bytes.Equal(computedChecksum, secret.Checksum[:]) {
		err = ErrInvalidPassword
		return
	}

	// Return result
	copy(data.Buffer[:], buffer[1:])
	data.Length = buffer[0]

	return
}
