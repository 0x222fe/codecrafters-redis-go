package state

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type UserFlag string

const (
	FlagNoPass UserFlag = "nopass"
)

var (
	DefaulUser = &User{
		name: "default",
		flags: map[UserFlag]any{
			FlagNoPass: true,
		},
		passwords: map[string]any{},
		disabled:  false,
	}
)

type User struct {
	mu        sync.RWMutex
	name      string
	flags     map[UserFlag]any
	passwords map[string]any
	disabled  bool
}

func NewUser(name string) *User {
	return &User{
		mu:        sync.RWMutex{},
		name:      name,
		flags:     map[UserFlag]any{},
		passwords: map[string]any{},
		disabled:  false,
	}
}

func (u *User) GetName() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.name
}

func (u *User) GetFlags() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()

	flags := make([]string, 0, len(u.flags))
	for flag := range u.flags {
		flags = append(flags, string(flag))
	}
	return flags
}

func (u *User) GetPasswords() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()

	passwords := make([]string, 0, len(u.passwords))
	for password := range u.passwords {
		passwords = append(passwords, password)
	}
	return passwords
}

func (u *User) AddPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	hashHex := hex.EncodeToString(hash[:])

	u.mu.Lock()
	defer u.mu.Unlock()

	u.passwords[hashHex] = true
	delete(u.flags, FlagNoPass)
	return hashHex
}

func (u *User) ValidatePassword(password string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.disabled {
		return false
	}

	if _, ok := u.flags[FlagNoPass]; ok {
		return true
	}

	hash := sha256.Sum256([]byte(password))
	hashHex := hex.EncodeToString(hash[:])
	_, ok := u.passwords[hashHex]
	return ok
}