package auth

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

type TokenManager struct {
    secret []byte
}

func NewTokenManager(secret string) *TokenManager {
    return &TokenManager{secret: []byte(secret)}
}

func GenerateRandomSecret() string {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "development-secret"
    }
    return base64.RawURLEncoding.EncodeToString(b)
}

type Claims struct {
    Sub  string `json:"sub"`
    Role string `json:"role"`
    Exp  int64  `json:"exp"`
}

func (t *TokenManager) Sign(userID, role string, ttl time.Duration) (string, error) {
    header := map[string]string{"alg": "HS256", "typ": "JWT"}
    hB, _ := json.Marshal(header)
    claims := Claims{Sub: userID, Role: role, Exp: time.Now().Add(ttl).Unix()}
    cB, _ := json.Marshal(claims)
    hEnc := base64.RawURLEncoding.EncodeToString(hB)
    cEnc := base64.RawURLEncoding.EncodeToString(cB)
    unsigned := hEnc + "." + cEnc
    sig := signHS256(unsigned, t.secret)
    return unsigned + "." + sig, nil
}

func (t *TokenManager) Verify(token string) (*Claims, error) {
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return nil, fmt.Errorf("invalid token format")
    }
    unsigned := parts[0] + "." + parts[1]
    if !verifyHS256(unsigned, parts[2], t.secret) {
        return nil, fmt.Errorf("invalid signature")
    }
    // decode claims
    cb, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return nil, fmt.Errorf("invalid claims encoding")
    }
    var c Claims
    if err := json.Unmarshal(cb, &c); err != nil {
        return nil, fmt.Errorf("invalid claims")
    }
    if time.Now().Unix() > c.Exp {
        return nil, fmt.Errorf("token expired")
    }
    return &c, nil
}

func signHS256(msg string, secret []byte) string {
    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(msg))
    sig := mac.Sum(nil)
    return base64.RawURLEncoding.EncodeToString(sig)
}

func verifyHS256(msg, sig string, secret []byte) bool {
    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(msg))
    s := mac.Sum(nil)
    got, err := base64.RawURLEncoding.DecodeString(sig)
    if err != nil || len(got) != len(s) {
        return false
    }
    var diff byte
    for i := range got {
        diff |= got[i] ^ s[i]
    }
    return diff == 0
}

