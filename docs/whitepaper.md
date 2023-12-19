# Self-hosted secret manager

Antonio Jose Donis Hung - antoniojosedonishung@gmail.com

## Abstract

The primary purpose of a password manager is to act as a digital memory aid, storing complex passwords that are otherwise challenging for most people to remember. Recognizing the necessity of password management, this document aims to address secondary concerns: What if I lose my phone? How can I access backups in emergencies? Does key based secrets are the only critical information to save? What about other files? Are my secrets inter-operable across different platforms and programs? Most critically, can I trust an online password manager with my sensitive data? This document outlines a system, not just a software, enabling anyone, regardless of their programming knowledge, to securely store passwords and other secrets on their own devices.

## Problem statement

The modern internet user, faces a significant issue: lack of digital sovereignty. Most rely on third-party companies for managing their online identities and assets. For instance, Google, a popular service for email setup, becomes a single point of failure; losing access to one's Google account can cascade into losing access to various services linked to that email. Similarly, when considering file backups, the go-to solutions are Google Drive, OneDrive, and iCloud – all controlled by major corporations. How can we be certain they can't access our files? Are they sharing my data with a nation-state? This uncertainty extends to online password managers, which might pose risks, especially for activists in hostile environments.

Key concerns include:

- The highly likely illegal miss-use of our data by companies and nation state driven adversaries.
- Lack of Zero-trust systems so end users don't need to trust any party for storing critical information.

## Philosophy

Inspired by the UNIX philosophy, zero trust principles, and decentralized foundations, this system adheres to the following constraints:

- **Secret**: A secret is just any blob of data in plain text or binary that an end user, company or organization wants to maintain confidential against domestic or military targeted espionage.
- **Privacy**: Secrets are only written to disks only if they have already been encrypted.
- **Portability**: The system avoids custom binary formats; it uses a JSON file as database, facilitating compatibility and transition between different tools. This approach addresses the common issue of proprietary, non-free formats prevalent in many mobile apps, which are challenging to reverse-engineer.
- **Decentralized**: Backup processes are founded on zero-trust principles, ensuring that no reliance is placed on the service provider for privacy or data longevity.

## Encryption

The system should be secured by mathematically proved cryptographic algorithm, combining multiple concepts such as **key derivation**, **keyed-hash message authentication code**, **AES CBC** and **zero obscurity**. With this key aspects the system should operate as it is, having its security in the robustness of the algorithm, rather than the obscurity to maintain secret how to interpret the finally stored bytes.

In this section the paper will focus on describing first, separately the why of the chosen parts to finally, describe the overall algorithm.

### Key derivation

To protect keys from been brute-forced by adversaries and reused in other points in which they could be valid, the system employs  `argon2id` to derive the real key into a `256-bits` (or `32 bytes`) long.

The most important point to have in count for the end developer creating a fork of the tool for an specific platform, is to always maintain the most performance expensive settings for the era. This way even if the adversary is able to still the original encrypted secret or its backup. It will still be hard for average consumer and company targeted hardware to break the key.

> IMPORTANT: Take in mind that if a nation state, with the enough resources and time for persecuting someone "to the end of the world" they will probably achieve it. BUT, IF AND ONLY IF ENCRYPTION PASSWORD IS "EASIER" TO BRUTE-FORCE. Check this small section explaining some recommendations for preserving security even with these kind of adversaries

To maintain a robust key will be important to follow these recommendations or more strict ones. Take in mind that this is, what (my self) the author does:

- Keys of `128-bits` (or `16 bytes`) to `256-bits` (or `32 bytes`) long. Longer are better. But consider the UX for memorizing it.

- Only English characters. (This will prevent targeted cultural guessing. Check **Other ways to guess the message size** section for more details).

- Upper case characters. (At least 4?)

- Lower case characters. (At least 4?)

- Digits. (At least 4?)

- ASCII Special characters including white space. (At least 4?)

- Non real words: Do not use real words, or modified words with other characters that could imaginary sound like real English words. The random they sound the less like they will be present in a dictionary.

- Only preserve the generated master-key in your memory.

- Finite life keys. For only encrypting a certain amount of secrets with them. This will reduce by consequence the amount of samples an adversary has for analyzing. Making the analysis even harder.

### Keyed-hash message authentication code

To ensure the end-user always decrypts trusted data and preventing him for generating and accidental (adversary intended) malicious sample for an attacker to analyse. The system employs an **HMAC(SHA3_512)** checksum with a second `argon2id` derivation of the previously derived key using a newly generated random salt.

This new key is used in hand with the actual cipher text to validate the authenticity and the authorship of the decryption key over the cipher. 

### AES CBC

Encryption is performed into a padded blocks with a length multiple of **256 bytes** long.
The padding is performed just in the latest block.

Padding is assigned from a cryptographic secure random source and the last byte of the last block is used to determine the number of bytes to ignore from the most right.

```
+-+-+-+-+-+-+--------------------------+-------------------+
|S|E|C|R|E|T|  PAD UNTIL LENGTH IS 255 | 255 - LEN(SECRET) |
+-+-+-+-+-+-+--------------------------+-------------------+
```

Prevention of **Oracle padding attack** is explained in more detail in the section with the same name.

### Zero obscurity

The main idea is. Even if the attackers knows the settings of `argon2id` the final **HMAC** checksum and the IV for **AES CBC** the system should be strong enough (considering the end user follows, the previous suggestions or stricter ones). Allowing the end user storage the final `argon2id`, **HMAC** and cipher text in any remote storage, trusted or not. To ensure this, the final JSON will maintain these properties:

```json
{
    "argon": {
        "time":    65536,
        "memory":  1024,
        "threads": 64
    },
    "iv":       [1, 2, 3, "...", 16],
    "keySalt":  [1, "...", "length-is-user-defined-during-encryption"],
    "hmacSalt": [1, "...", "length-is-user-defined-during-encryption"],
    "cipher":   ["Array with a length multiple of 256 bytes long"],
    "hmac":     [1, "...", 512]
}
```

Argon attributes are: **time** the number of passes to perform. **memory** in **KiB** and **threads** is the number of CPU threads to actually use.

### The algorithm

#### Encryption

```go
// Prepares a secret structure with a ready to use IV and salts
func (j *Job) Encrypt() (secret *Secret) {
    dataLength := ChunkSize * (1 + len(j.Data)/ChunkSize)
    secret = &Secret{
        Argon:    j.Argon,
        IV:       make([]byte, IVSize),
        KeySalt:  make([]byte, j.SaltSize),
        HMACSalt: make([]byte, j.SaltSize),
        Cipher:   make([]byte, dataLength),
    }
    // Prepare random data
    rand.Read(secret.IV)
    rand.Read(secret.KeySalt)
    rand.Read(secret.HMACSalt)

    // Prepare data to encrypt
    data := make([]byte, dataLength)
    defer rand.Read(data)
    copy(data, j.Data)
    rand.Read(data[len(j.Data):])
    data[dataLength-1] = byte(dataLength - len(j.Data))

    // Prepare encryption key
    key := argon2.IDKey(j.Key, secret.KeySalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, KeySize)

    // Encrypt data
    // Error doesn't need verification because key is always of valid size, thanks to argon
    block, _ := aes.NewCipher(key)
    enc := cipher.NewCBCEncrypter(block, secret.IV)
    enc.CryptBlocks(secret.Cipher, data)

    // Calculate HMAC sum
    hmacKey := argon2.IDKey(key, secret.HMACSalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, HMACKeySize)
    hash := hmac.New(sha3.New512, hmacKey)
    hash.Write(secret.Cipher)
    secret.HMAC = hash.Sum(nil)

    return secret
}
```

#### Decryption

```go
// Decrypt populates the Data field of Job struct with the decrypted secret on success
// On failure returns ErrDecryptionFailed
func (j *Job) Decrypt(secret *Secret) (err error) {
	// Prepare decryption key
	key := argon2.IDKey(j.Key, secret.KeySalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, KeySize)

	// Verify HMAC
	hmacKey := argon2.IDKey(key, secret.HMACSalt, secret.Argon.Time, secret.Argon.Memory, secret.Argon.Threads, HMACKeySize)
	hash := hmac.New(sha3.New512, hmacKey)
	hash.Write(secret.Cipher)
	computedHMAC := hash.Sum(nil)
	if !bytes.Equal(secret.HMAC, computedHMAC) {
		return ErrDecryptionFailed
	}

	// Prepare decrypt buffer
	data := make([]byte, len(secret.Cipher))
	defer rand.Read(data)

	// Decrypt data
	block, _ := aes.NewCipher(key)
	enc := cipher.NewCBCDecrypter(block, secret.IV)
	enc.CryptBlocks(data, secret.Cipher)

	// Copy Data
	realLength := len(data) - int(data[len(data)-1])
	j.Data = make([]byte, realLength)
	copy(j.Data, data[:realLength])

	return err
}
```

### Ideas of attacks and their prevention

This section addresses potential security concerns to demonstrate the robustness of the cryptographic approach in the system.

#### Encryption key protection

To safeguard against the unauthorized extraction of the encryption key, which could lead to the decryption of all stored keys, the system employs key stretching through `argon2id`. This approach generates a distinct key for each stored secret, diverging the resulted cipher-text even more. Each secret's encryption key is computed individually, compelling an attacker to break into each one separately. If an attacker targets the encryption key, they will face a significantly more complex time challenge, since they will to generate a arbitrary length password of which then compute the `argon2id`. This increased security level is contingent upon the `argon2id` parameters being set to create a computationally demanding function. To maximize security, applications using this system are **advised to configure these parameters optimally**, tailored to the specific hardware capabilities of each platform in hand of using a strong key.

#### Oracle padding attack

Oracle padding attack consist in a malicious actor studying how the encryption and decryption algorithm behave against invalid input. This attack focus on the padding calculation part, of which the attacker expects the cryptographic algorithm to crash or respond with an `Invalid padding` error. This will let the attacker ex-filtrate the message size, which then will allow him to easy the process of breaking the final cipher.

`guardian` prevents this by using fixed **256** bytes long blocks. Ff which, the last block includes the random padding. Having the last bytes correspond to the number of bytes to ignore of this final block. Since an `uint8` has a minimum value of **0** and a maximum value of **255**. The resulting padding size will be always inside the block size. Making impossible for an attacker to perform a padding guessing using indexing errors.

#### Other ways to guess the message size

An attacker could perform a message size guessing by reading until a non printable character is reached. In the example bellow the user encrypted the secret **KEY** of which the vanilla random padding generated a buffer of which the first byte is a non printable character `\n`. Allowing the attacker perform targeted detection in case the secret was not randomly generated. Like in this example, by having the secret be a valid English word.

```
+------------+--------+--------+--------+-------------+
| Msg lengtg | Byte 1 | Byte 2 | Byte 3 | Padding ... |
+------------+--------+--------+--------+-------------+
|     3      |    K   |   E    |    Y   |   0x0A      |
+------------+--------+--------+--------+-------------+
```

This could be prevented by only using the same set of characters for the actual key and the secret to encrypt. And writing a custom random source that filters only printable ASCII characters from the final cryptographically secure random source.

## File sharing

```
TODO: This section is still in design
```
