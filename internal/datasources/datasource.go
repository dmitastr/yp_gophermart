package datasources

import (
	"github.com/dmitastr/yp_gophermart/internal/domain/models"
)

type Database interface {
	RegisterUser(models.User) error
	GetUser(string) (models.User, error)
}
