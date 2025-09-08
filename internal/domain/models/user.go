package models

import (
	"time"

	"github.com/jackc/pgx/v5"
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
