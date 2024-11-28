package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type User struct {
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	HomeDir  string `yaml:"homedir" mapstructure:"homedir"`
}

func (u *User) String() string {
	return fmt.Sprintf("%s:%s:%s", u.Username, u.Password, u.HomeDir)
}

type Users []User

func (u *Users) String() string {
	users := []string{}
	for _, user := range *u {
		users = append(users, user.String())
	}

	return strings.Join(users, ",")
}

func (u *Users) Set(value string) error {
	users := strings.Split(value, ",")
	for _, user := range users {
		values := strings.Split(user, ":")
		if len(values) != 3 {
			return fmt.Errorf("invalid user format: %s", value)
		}

		user := User{
			Username: values[0],
			Password: values[1],
			HomeDir:  values[2],
		}

		*u = append(*u, user)
	}
	return nil
}

func (u *Users) Type() string {
	return "user"
}

func (u *Users) FindByUsername(username string) (*User, bool) {
	for _, user := range *u {
		if user.Username == username {
			return &user, true
		}
	}

	return nil, false
}

func UserStringDecodeHook() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.String && to == reflect.TypeOf(Users{}) {
			users := Users{}
			if err := users.Set(data.(string)); err != nil {
				return nil, err
			}

			return users, nil
		}

		return data, nil
	}
}

type Config struct {
	RootDir        string `yaml:"root_dir"`
	Users          Users  `yaml:"users"`
	AllowAnonymous bool   `yaml:"allow_anonymous"`
	MinPassivePort int    `yaml:"min_passive_port"`
	MaxPassivePort int    `yaml:"max_passive_port"`
	Address        string `yaml:"address"`
	Port           int    `yaml:"port"`
}

func (c *Config) Validate() error {
	if c.RootDir == "" {
		return fmt.Errorf("root_dir is required")
	}

	if len(c.Users) == 0 && !c.AllowAnonymous {
		return fmt.Errorf("users or allow_anonymous is required")
	}

	if c.MinPassivePort >= c.MaxPassivePort {
		return fmt.Errorf("min_passive_port should be less than max_passive_port")
	}

	return nil
}
