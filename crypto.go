package jar

// ECB mode assumed Public Domain, modified from https://gist.github.com/DeanThompson/17056cc40b4899e3e7f4

import (
	httpauth "github.com/abbot/go-http-auth"

	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"fmt"
)

const (
	// ErrCiphertextTooShort is returned when the ciphertext is too damn short
	ErrCiphertextTooShort = Error("ciphertext too short")

	// ErrCiphertextIrregular is returned when the ciphertext is not a multiple of the block size
	ErrCiphertextIrregular = Error("ciphertext is not a multiple of the block size")

	// errMismatchedHashAndPassword is reimplemented from go-http-auth to support some
	// reimplemented-because-unexported functions
	errMismatchedHashAndPassword = Error("mismatched hash and password")
)

// Constants for configuration key strings
const (
	ConfigTLSCerts             = ConfigKey("tls.certs")
	ConfigTLSCiphers           = ConfigKey("tls.ciphers")
	ConfigTLSEnabled           = ConfigKey("tls.enabled")
	ConfigTLSHTTPRedirects     = ConfigKey("tls.httpredirects")
	ConfigTLSKeepaliveDisabled = ConfigKey("tls.keepalivedisabled")
	ConfigTLSListen            = ConfigKey("tls.listen")
	ConfigTLSMaxVersion        = ConfigKey("tls.maxversion")
	ConfigTLSMinVersion        = ConfigKey("tls.minversion")
	ConfigTLSHTTP2             = ConfigKey("tls.http2")
)

var (
	// Ciphers is a map of ciphers from crypto/tls
	Ciphers SuiteMap

	/* Legacy Ciphers definition kept for reference
	Ciphers = SuiteMap{
		"TLS_RSA_WITH_RC4_128_SHA":                0x0005,
		"TLS_RSA_WITH_3DES_EDE_CBC_SHA":           0x000a,
		"TLS_RSA_WITH_AES_128_CBC_SHA":            0x002f,
		"TLS_RSA_WITH_AES_256_CBC_SHA":            0x0035,
		"TLS_RSA_WITH_AES_128_CBC_SHA256":         0x003c,
		"TLS_RSA_WITH_AES_128_GCM_SHA256":         0x009c,
		"TLS_RSA_WITH_AES_256_GCM_SHA384":         0x009d,
		"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":        0xc007,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    0xc009,
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    0xc00a,
		"TLS_ECDHE_RSA_WITH_RC4_128_SHA":          0xc011,
		"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":     0xc012,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      0xc013,
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      0xc014,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256": 0xc023,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":   0xc027,
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   0xc02f,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": 0xc02b,
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   0xc030,
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": 0xc02c,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    0xcca8,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  0xcca9,
		// TLS 1.3 cipher suites.
		"TLS_AES_128_GCM_SHA256":       0x1301,
		"TLS_AES_256_GCM_SHA384":       0x1302,
		"TLS_CHACHA20_POLY1305_SHA256": 0x1303,
	}
	*/

	// SslVersions is a map of SSL/TLS versions, mapped locally
	SslVersions = SuiteMap{
		"VersionSSL30": 0x0300,
		"VersionTLS10": 0x0301,
		"VersionTLS11": 0x0302,
		"VersionTLS12": 0x0303,
		"VersionTLS13": 0x0304,
	}
)

func init() {

	Ciphers = NewSuiteMapFromCipherSuites(tls.CipherSuites())

	ConfigAdditions[ConfigTLSCiphers] = []string{
		// TLS 1.3
		"TLS_CHACHA20_POLY1305_SHA256",
		"TLS_AES_256_GCM_SHA384",
		"TLS_AES_128_GCM_SHA256",
		// TLS < 1.3
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		//"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
		//"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
	}
	ConfigAdditions[ConfigTLSMinVersion] = float64(1.2)
	ConfigAdditions[ConfigTLSMaxVersion] = float64(1.3)
	ConfigAdditions[ConfigTLSListen] = ":8443"
}

// Cert encapsulated a Domain, the Keyfile, and a Certfile
type Cert struct {
	Domain   string
	Keyfile  string
	Certfile string
}

// SuiteMap is a map of TLS cipher suites, to their hex code
type SuiteMap map[string]uint16

// NewSuiteMapFromCipherSuites takes a []*CipherSuite and creates a SuiteMap from it
func NewSuiteMapFromCipherSuites(cipherSuites []*tls.CipherSuite) SuiteMap {
	s := make(SuiteMap)
	for _, c := range cipherSuites {
		s[c.Name] = c.ID
	}

	// Accommodate legacy suite names, for backwards compatibility
	if _, ok := s["TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"]; ok {
		s["TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"] = s["TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"]
	}
	if _, ok := s["TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"]; ok {
		s["TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"] = s["TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"]
	}

	return s

}

// List returns the names of the cipher suites in an untrustable order
func (s *SuiteMap) List() []string {
	cl := make([]string, len(*s))
	i := 0
	for k := range *s {
		cl[i] = k
		i++
	}
	return cl
}

// AllSuites returns the hex codes for all of the cipher suites in an untrustable order
func (s *SuiteMap) AllSuites() []uint16 {
	c, _ := s.CipherListToSuites(s.List())
	return c
}

// CipherListToSuites takes an ordered list of cipher suite names, and returns their hex codes in the same order
func (s *SuiteMap) CipherListToSuites(list []string) ([]uint16, error) {
	ilist := make([]uint16, len(list))
	m := *s
	for i, cipher := range list {
		if v, ok := m[cipher]; ok {
			ilist[i] = v
		} else {
			return []uint16{}, fmt.Errorf("cipher '%s' doesn't exist", cipher)
		}
	}
	return ilist, nil
}

// Suite reverse lookups a suitename given the number
func (s *SuiteMap) Suite(number uint16) string {
	for k, v := range *s {
		if v == number {
			return k
		}
	}
	return ""
}

// ECBDecrypt takes a base64-encoded key and RawURLencoded-base64 ciphertext to decrypt, and returns the plaintext or an error.
// PKCS5 padding is trimmed as needed
func ECBDecrypt(b64key string, eb64ciphertext string) (plaintext []byte, err error) {
	key, _ := base64.StdEncoding.DecodeString(b64key)

	ciphertext, err := base64.RawURLEncoding.DecodeString(eb64ciphertext)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	if len(ciphertext) < aes.BlockSize {
		err = ErrCiphertextTooShort
		return
	}
	//ciphertext = ciphertext[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		err = ErrCiphertextIrregular
		return
	}

	mode := NewECBDecrypter(block)
	mode.CryptBlocks(ciphertext, ciphertext)

	return trimPKCS5(ciphertext), nil
}

// ECBEncrypt takes a base64-encoded key and a []byte, and returns the base64-encdoded ciphertext or an error.
// PKCS5 padding is added as needed
func ECBEncrypt(b64key string, plaintext []byte) (b64ciphertext string, err error) {
	key, _ := base64.StdEncoding.DecodeString(b64key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	plaintext = addPKCS5(plaintext, aes.BlockSize)

	mode := NewECBEncrypter(block)
	mode.CryptBlocks(plaintext, plaintext)

	b64ciphertext = base64.RawURLEncoding.EncodeToString(plaintext)
	return
}

type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

// NewECBEncrypter should never be used unless you know what you're doing
func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

// BlockSize returns the blocksize
func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

// CryptBlocks encrypts blocks
func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		ErrorOut.Println("ecb error: crypto/cipher: input not full blocks")
		return
	}
	if len(dst) < len(src) {
		ErrorOut.Println("ecb error: crypto/cipher: output smaller than input")
		return
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

// NewECBDecrypter should never be used unless you know what you're doing
func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

// BlockSize returns the blocksize
func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

// CryptBlocks encrypts blocks
func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		ErrorOut.Println("ecb error: crypto/cipher: input not full blocks")
		return
	}
	if len(dst) < len(src) {
		ErrorOut.Println("ecb error: crypto/cipher: output smaller than input")
		return
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// trimPKCS5 trims padding based on PKCS5 rules
func trimPKCS5(plaintext []byte) []byte {
	padding := plaintext[len(plaintext)-1]
	return plaintext[:len(plaintext)-int(padding)]
}

// addPKCS5 adds padding based on PKCS5 rules
func addPKCS5(plaintext []byte, blocksize int) []byte {
	padding := blocksize - len(plaintext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

// compareShaHashAndPassword compares a SHA-hashed password and a clear password and returns errMismatchedHashAndPassword if they do not match
func compareShaHashAndPassword(hashedPassword, password []byte) error {
	d := sha1.New()
	d.Write(password)
	if subtle.ConstantTimeCompare(hashedPassword[5:], []byte(base64.StdEncoding.EncodeToString(d.Sum(nil)))) != 1 {
		return errMismatchedHashAndPassword
	}
	return nil
}

// compareMD5HashAndPassword compares an MD5-hashed password and a clear password and returns errMismatchedHashAndPassword if they do not match
func compareMD5HashAndPassword(hashedPassword, password []byte) error {
	parts := bytes.SplitN(hashedPassword, []byte("$"), 4)
	if len(parts) != 4 {
		return errMismatchedHashAndPassword
	}
	magic := []byte("$" + string(parts[1]) + "$")
	salt := parts[2]
	if subtle.ConstantTimeCompare(hashedPassword, httpauth.MD5Crypt(password, salt, magic)) != 1 {
		return errMismatchedHashAndPassword
	}
	return nil
}
