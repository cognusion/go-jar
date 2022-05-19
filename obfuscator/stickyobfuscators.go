package obfuscator

import (
	"github.com/cognusion/oxy/roundrobin"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(ioutil.Discard, "[DEBUG] ", 0)
	// ErrorOut is a log.Logger for error messages
	ErrorOut = log.New(ioutil.Discard, "", 0)
)

// AesObfuscator is a roundrobin.Obfuscator that returns an nonceless encrypted version
type AesObfuscator struct {
	block cipher.AEAD
	ttl   time.Duration
}

// NewAesObfuscator takes a fixed-size key and returns an Obfuscator or an error.
// Key size must be exactly one of 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewAesObfuscator(key []byte) (roundrobin.Obfuscator, error) {
	var a AesObfuscator

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	a.block = aesgcm

	return &a, nil
}

// NewAesObfuscatorWithExpiration takes a fixed-size key and a TTL, and returns an Obfuscator or an error.
// Key size must be exactly one of 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewAesObfuscatorWithExpiration(key []byte, ttl time.Duration) (roundrobin.Obfuscator, error) {
	var a AesObfuscator

	a.ttl = ttl
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	a.block = aesgcm

	return &a, nil
}

// Obfuscate takes a raw string and returns the obfuscated value
func (o *AesObfuscator) Obfuscate(raw string) string {
	if o.ttl > 0 {
		raw = fmt.Sprintf("%s|%d", raw, time.Now().UTC().Add(o.ttl).Unix())
	}

	/*
		Nonce is the 64bit nanosecond-resolution time, plus 32bits of crypto/rand, for 96bits (12Bytes).
		Theoretically, if 2^32 calls were made in 1 nanosecon, there might be a repeat.
		Adds ~765ns, and 4B heap in 1 alloc (Benchmark_NonceTimeRandom4 below)

		Benchmark_NonceRandom12-8      	 2000000	       723 ns/op	      16 B/op	       1 allocs/op
		Benchmark_NonceRandom4-8       	 2000000	       698 ns/op	       4 B/op	       1 allocs/op
		Benchmark_NonceTimeRandom4-8   	 2000000	       765 ns/op	       4 B/op	       1 allocs/op
	*/
	nonce := make([]byte, 12)
	binary.PutVarint(nonce, time.Now().UnixNano())
	rpend := make([]byte, 4)
	if _, err := io.ReadFull(rand.Reader, rpend); err != nil {
		// This is a near-impossible error condition on Linux systems.
		// An error here means rand.Reader (and thus getrandom(2), and thus /dev/urandom) returned
		// less than 4 bytes of data. /dev/urandom is guaranteed to always return the number of
		// bytes requested up to 512 bytes on modern kernels. Behaviour on non-Linux systems
		// varies, of course.
		ErrorOut.Printf("AesObfuscator.Obfuscate randReader error: %s. Panicing\n", err.Error())
		DebugOut.Printf("AesObfuscator.Obfuscate randReader error: %s. Panicing\n", err.Error())
		panic(err)
	}
	for i := 0; i < 4; i++ {
		nonce[i+8] = rpend[i]
	}

	obfuscated := o.block.Seal(nil, nonce, []byte(raw), nil)
	// We append the 12byte nonce onto the end of the message
	obfuscated = append(obfuscated, nonce...)
	obfuscatedStr := base64.RawURLEncoding.EncodeToString(obfuscated)
	return obfuscatedStr
}

// Normalize takes an obfuscated string and returns the raw value
func (o *AesObfuscator) Normalize(obfuscatedStr string) string {
	obfuscated, err := base64.RawURLEncoding.DecodeString(obfuscatedStr)
	if err != nil {
		ErrorOut.Printf("AesObfuscator.Normalize Decoding base64 failed with '%s'\n", err)
		return ""
	}

	// The first len-12 bytes is the ciphertext, the last 12 bytes is the nonce
	n := len(obfuscated) - 12
	if n <= 0 {
		// Protect against range errors causing panics
		ErrorOut.Printf("AesObfuscator.Normalize post-base64-decoded string is too short\n")
		return ""
	}

	nonce := obfuscated[n:]
	obfuscated = obfuscated[:n]

	raw, err := o.block.Open(nil, nonce, []byte(obfuscated), nil)
	if err != nil {
		// um....
		ErrorOut.Printf("AesObfuscator.Normalize Open failed with '%s'\n", err)
		return "" // (badpokerface)
	}
	if o.ttl > 0 {
		rawparts := strings.Split(string(raw), "|")
		if len(rawparts) < 2 {
			ErrorOut.Printf("AesObfuscator.Normalize TTL set but cookie doesn't contain an expiration: '%s'\n", raw)
			return "" // (sadpanda)
		}
		// validate the ttl
		i, err := strconv.ParseInt(rawparts[1], 10, 64)
		if err != nil {
			ErrorOut.Printf("AesObfuscator.Normalize TTL can't be parsed: '%s'\n", raw)
			return "" // (sadpanda)
		}
		if time.Now().UTC().After(time.Unix(i, 0).UTC()) {
			strTime := time.Unix(i, 0).UTC().String()
			ErrorOut.Printf("AesObfuscator.Normalize TTL expired: '%s' (%s)\n", raw, strTime)
			DebugOut.Printf("AesObfuscator.Normalize TTL expired: '%s' (%s)\n", raw, strTime)
			return "" // (curiousgeorge)
		}
		raw = []byte(rawparts[0])
	}
	return string(raw)
}

// HexObfuscator is a roundrobin.Obfuscator that returns an hex-encoded version of the value
type HexObfuscator struct{}

// Obfuscate takes a raw string and returns the obfuscated value
func (o *HexObfuscator) Obfuscate(raw string) string {
	return hex.EncodeToString([]byte(raw))
}

// Normalize takes an obfuscated string and returns the raw value
func (o *HexObfuscator) Normalize(obfuscatedStr string) string {
	clear, _ := hex.DecodeString(obfuscatedStr)
	return string(clear)
}
