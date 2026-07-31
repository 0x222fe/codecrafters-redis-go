package state

import "github.com/0x222fe/codecrafters-redis-go/internal/user"

func (s *AppState) GetUser(name string) (*user.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[name]
	return user, exists
}

func (s *AppState) AddUser(user *user.User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.Name()] = user
}
