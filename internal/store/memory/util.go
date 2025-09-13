package memory

import (
    "fmt"
    "sync"
    "time"
)

var (
    counterMutex sync.Mutex
    lastTime     int64
    counter      int
)

func newID() string {
    counterMutex.Lock()
    defer counterMutex.Unlock()
    
    now := time.Now().Unix()
    if now != lastTime {
        lastTime = now
        counter = 1
    } else {
        counter++
    }
    
    return fmt.Sprintf("%d_%04d", now, counter)
}

// NewIDForExternal exposes ID generation for other stores.
func NewIDForExternal() string { return newID() }
