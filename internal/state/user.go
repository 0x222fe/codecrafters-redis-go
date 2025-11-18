package state

type UserFlag string

const (
	FlagNoPass UserFlag = "nopass"
)

type User struct {
	Name  string
	Flags map[UserFlag]any
}

func NewUser(name string, flags []UserFlag) *User {
	flagMap := make(map[UserFlag]any, len(flags))
	for _, f := range flags {
		flagMap[UserFlag(f)] = true
	}
	return &User{
		Name:  name,
		Flags: flagMap,
	}
}

func (u *User) GetFlags() []string {
	flags := make([]string, 0, len(u.Flags))
	for flag := range u.Flags {
		flags = append(flags, string(flag))
	}
	return flags
}
