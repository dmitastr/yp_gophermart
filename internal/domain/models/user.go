package models

import (
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name     string `json:"login"`
	Password string `json:"password"`
	Hash     string `json:"hash,omitempty"`
}

func (u *User) ToNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{"name": u.Name, "pass": u.Hash, "created_at": time.Now()}
}

func (u *User) IsValid() bool {
	return u.Name != "" && u.Password != ""
}

func (u *User) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Hash = string(hash)
	return nil
}
