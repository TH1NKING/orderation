package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
)

// PasswordHasher provides a lightweight, iterative salted hash.
// NOTE: For production, swap to a proven KDF like bcrypt, scrypt or argon2.
type PasswordHasher struct {
    iterations int
}

func NewPasswordHasher(iterations int) *PasswordHasher {
    if iterations <= 0 {
        iterations = 200_000
    }
    return &PasswordHasher{iterations: iterations}
}

func (p *PasswordHasher) Hash(plain string) (string, error) {
    salt := make([]byte, 16)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }
    sum := sha256.Sum256(append(salt, []byte(plain)...))
    h := sum[:]
    for i := 0; i < p.iterations; i++ {
        s := sha256.Sum256(h)
        h = s[:]
    }
    // format: iterations$salt$b64hash
    return encodeInt(p.iterations) + "$" + base64.RawURLEncoding.EncodeToString(salt) + "$" + base64.RawURLEncoding.EncodeToString(h), nil
}

func (p *PasswordHasher) Verify(plain, encoded string) bool {
    // parse
    parts := split3(encoded)
    if parts[0] == "" || parts[1] == "" || parts[2] == "" {
        return false
    }
    iter := decodeInt(parts[0])
    salt, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return false
    }
    want, err := base64.RawURLEncoding.DecodeString(parts[2])
    if err != nil {
        return false
    }
    sum := sha256.Sum256(append(salt, []byte(plain)...))
    h := sum[:]
    for i := 0; i < iter; i++ {
        s := sha256.Sum256(h)
        h = s[:]
    }
    if len(want) != len(h) {
        return false
    }
    // constant-time compare
    var diff byte
    for i := range want {
        diff |= want[i] ^ h[i]
    }
    return diff == 0
}

func split3(s string) [3]string {
    var a, b, c string
    i := indexByte(s, '$')
    if i < 0 {
        return [3]string{}
    }
    a = s[:i]
    s = s[i+1:]
    j := indexByte(s, '$')
    if j < 0 {
        return [3]string{}
    }
    b = s[:j]
    c = s[j+1:]
    return [3]string{a, b, c}
}

func indexByte(s string, c byte) int {
    for i := 0; i < len(s); i++ {
        if s[i] == c {
            return i
        }
    }
    return -1
}

func encodeInt(n int) string {
    // encode base64 url
    b := []byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
    return base64.RawURLEncoding.EncodeToString(b)
}

func decodeInt(s string) int {
    b, err := base64.RawURLEncoding.DecodeString(s)
    if err != nil || len(b) != 4 {
        return 0
    }
    return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}

