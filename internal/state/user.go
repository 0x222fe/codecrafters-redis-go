package state

type User struct {
	Name  string
	Flags []string
}

func NewUser(name string) *User {
	return &User{
		Name:  name,
		Flags: []string{},
	}
}