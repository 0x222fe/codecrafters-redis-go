package state

type UserFlag string

const (
	FlagNoPass UserFlag = "nopass"
)

var (
	DefaulUser = &User{
		Name: "default",
		Flags: map[UserFlag]any{
			FlagNoPass: true,
		},
		Passwords: map[string]any{},
	}
)

type User struct {
	Name      string
	Flags     map[UserFlag]any
	Passwords map[string]any
}

func (u *User) GetInfo() map[string][]string {
	flags := make([]string, 0, len(u.Flags))
	for flag := range u.Flags {
		flags = append(flags, string(flag))
	}

	passwords := make([]string, 0, len(u.Passwords))
	for password := range u.Passwords {
		passwords = append(passwords, password)
	}

	return map[string][]string{
		"flags":     flags,
		"passwords": passwords,
	}

}

// func (u *User) GetFlags() []string {
// 	flags := make([]string, 0, len(u.Flags))
// 	for flag := range u.Flags {
// 		flags = append(flags, string(flag))
// 	}
// 	return flags
// }
//
// func (u *User) GetPasswords() []string {
// 	passwords := make([]string, 0, len(u.Passwords))
// 	for password := range u.Passwords {
// 		passwords = append(passwords, password)
// 	}
// 	return passwords
// }
