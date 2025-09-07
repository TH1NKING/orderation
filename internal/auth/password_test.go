package auth

import "testing"

func TestPasswordHashAndVerify(t *testing.T) {
    h := NewPasswordHasher(500) // faster for tests
    hash1, err := h.Hash("secret")
    if err != nil {
        t.Fatalf("hash error: %v", err)
    }
    if !h.Verify("secret", hash1) {
        t.Fatalf("expected verify ok for correct password")
    }
    if h.Verify("wrong", hash1) {
        t.Fatalf("expected verify false for wrong password")
    }
    hash2, err := h.Hash("secret")
    if err != nil {
        t.Fatalf("hash2 error: %v", err)
    }
    if hash1 == hash2 {
        t.Fatalf("expected different hashes due to random salt")
    }
}

