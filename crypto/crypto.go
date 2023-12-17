package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/sha3"
)

var (
	ErrDecryptionFailed = errors.New("decryption failed")
)

const (
	ChecksumSize = 512 / 8
	ChunkSize    = 256
	SaltSize     = 512
	IVSize       = aes.BlockSize
	DataSize     = ChunkSize - 1
	KeySize      = 32
)

type Secret struct {
	// Configuration for the argon function
	Argon Argon `json:"argon"`

	// Initialization Vector (IV)
	// To initialize it always all the Init() function
	IV []byte `json:"iv"`

	// Key salt is the Argon Salt to use for the key stretching process
	// To initialize it always all the Init() function
	KeySalt []byte `json:"keySalt"`

	// Checksum salt is the Argon Salt to use for the checksum stretching process
	// To initialize it always all the Init() function
	ChecksumSalt []byte `json:"checksumSalt"`

	// Salt used for the HMAC calculation
	HMACSalt []byte `json:"hmacSalt"`

	// Checksum is 512 bits (256 bytes) long Cryptographic checksum
	// Algorithm used is SHA3-512
	// The checksum is computed for the output of the: stretch_cipher = Argon(cipher, salt) -> sha3_512(stretch_cipher)
	// This checksum is used to verify the decrypted buffer was correct
	Checksum []byte `json:"checksum"`

	// The actual cipher text created after encrypting the msg
	// Divided in blocks of 256 bytes
	// The last block correspond to the actual data and a padding. Of which to prevent Padding oracle attack
	// The padding is always a valid number
	Cipher []byte `json:"cipher"`

	// HMAC is used to verify the authenticity of the encrypted cipher
	// This will ensure the algorithm enver tries to decrypt data user never encrypted
	HMAC []byte `json:"hmac"`
}

func (s *Secret) Release() {
	s.Argon.Release()
	rand.Read(s.IV)
	rand.Read(s.KeySalt)
	rand.Read(s.ChecksumSalt)
	rand.Read(s.HMACSalt)
	rand.Read(s.Checksum)
	rand.Read(s.Cipher)
	rand.Read(s.HMAC)
}

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

type Job struct {
	Key, Data []byte
	Argon     Argon
	SaltSize  int
}

func (j *Job) Release() {
	rand.Read(j.Key)
	rand.Read(j.Data)
	j.Argon.Release()
	j.SaltSize = 0
}

// Prepares a secret structure with a ready to use IV and salts
func (j *Job) Encrypt() (secret *Secret) {
	dataLength := ChunkSize * (1 + len(j.Data)/ChunkSize)
	secret = &Secret{
		Argon:        j.Argon,
		IV:           make([]byte, IVSize),
		KeySalt:      make([]byte, j.SaltSize),
		ChecksumSalt: make([]byte, j.SaltSize),
		HMACSalt:     make([]byte, j.SaltSize),
		Cipher:       make([]byte, dataLength),
	}
	// Prepare random data
	rand.Read(secret.IV)
	rand.Read(secret.KeySalt)
	rand.Read(secret.ChecksumSalt)
	rand.Read(secret.HMACSalt)

	// Prepare data to encrypt
	data := make([]byte, dataLength)
	defer rand.Read(data)
	copy(data, j.Data)
	rand.Read(data[len(j.Data):])
	data[dataLength-1] = byte(dataLength - len(j.Data))

	// Calculate checksum
	rawChecksum := argon2.IDKey(data, secret.ChecksumSalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, ChecksumSize)
	hash := sha3.New512()
	hash.Write(rawChecksum)
	secret.Checksum = hash.Sum(nil)

	// Prepare encryption key
	key := argon2.IDKey(j.Key, secret.KeySalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, KeySize)

	// Encrypt data
	// Error doesn't need verification because key is always of valid size, thanks to argon
	block, _ := aes.NewCipher(key)
	enc := cipher.NewCBCEncrypter(block, secret.IV)
	enc.CryptBlocks(secret.Cipher, data)

	// Calculate HMAC sum
	hmacKey := argon2.IDKey(j.Key, secret.HMACSalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, KeySize)
	hash = hmac.New(sha3.New512, hmacKey)
	hash.Write(secret.Cipher)
	secret.HMAC = hash.Sum(nil)

	return secret
}

// Decrypt populates the Data field of Job struct with the decrypted secret on success
// On failure returns ErrDecryptionFailed
func (j *Job) Decrypt(secret *Secret) (err error) {
	// Verify HMAC
	hmacKey := argon2.IDKey(j.Key, secret.HMACSalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, KeySize)
	hash := hmac.New(sha3.New512, hmacKey)
	hash.Write(secret.Cipher)
	computedHMAC := hash.Sum(nil)
	if !bytes.Equal(secret.HMAC, computedHMAC) {
		return ErrDecryptionFailed
	}

	// Prepare decrypt buffer
	data := make([]byte, len(secret.Cipher))
	defer rand.Read(data)

	// Prepare decryption key
	key := argon2.IDKey(j.Key, secret.KeySalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, KeySize)

	// Decrypt data
	block, _ := aes.NewCipher(key)
	enc := cipher.NewCBCDecrypter(block, secret.IV)
	enc.CryptBlocks(data, secret.Cipher)

	// Verify checksum
	rawChecksum := argon2.IDKey(data, secret.ChecksumSalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, ChecksumSize)
	hash = sha3.New512()
	hash.Write(rawChecksum)
	computedChecksum := hash.Sum(nil)
	if !bytes.Equal(secret.Checksum, computedChecksum) {
		return ErrDecryptionFailed
	}

	// Copy Data
	realLength := len(data) - int(data[len(data)-1])
	j.Data = make([]byte, realLength)
	copy(j.Data, data[:realLength])

	return err
}
