package auth

import (
    "strings"
    "testing"
    "time"
)

func TestTokenSignAndVerify(t *testing.T) {
    tm := NewTokenManager("testsecret")
    tok, err := tm.Sign("u1", "user", time.Second)
    if err != nil {
        t.Fatalf("sign error: %v", err)
    }
    c, err := tm.Verify(tok)
    if err != nil {
        t.Fatalf("verify error: %v", err)
    }
    if c.Sub != "u1" || c.Role != "user" {
        t.Fatalf("unexpected claims: %+v", c)
    }

    // expired
    expired, _ := tm.Sign("u1", "user", -1*time.Second)
    if _, err := tm.Verify(expired); err == nil {
        t.Fatalf("expected expired token to fail verify")
    }

    // tamper signature
    parts := strings.Split(tok, ".")
    parts[2] = parts[2][:len(parts[2])-1] + "x"
    bad := strings.Join(parts, ".")
    if _, err := tm.Verify(bad); err == nil {
        t.Fatalf("expected tampered token to fail verify")
    }
}

