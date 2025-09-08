package memory

import (
    "crypto/rand"
    "encoding/hex"
)

func newID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}

// NewIDForExternal exposes ID generation for other stores.
func NewIDForExternal() string { return newID() }
