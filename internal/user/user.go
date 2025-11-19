package user

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type UserFlag string

const (
	FlagNoPass UserFlag = "nopass"
)

const (
	DefaultUserName = "default"
)

type User struct {
	mu        sync.RWMutex
	name      string
	flags     map[UserFlag]any
	passwords map[string]any
	disabled  bool
}

func New(name string) *User {
	return &User{
		mu:        sync.RWMutex{},
		name:      name,
		flags:     map[UserFlag]any{},
		passwords: map[string]any{},
		disabled:  false,
	}
}

func (u *User) Name() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.name
}

func (u *User) Flags() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()

	flags := make([]string, 0, len(u.flags))
	for flag := range u.flags {
		flags = append(flags, string(flag))
	}
	return flags
}

func (u *User) AddFlag(flag UserFlag) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.flags[flag] = true
}

func (u *User) Passwords() []string {
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
