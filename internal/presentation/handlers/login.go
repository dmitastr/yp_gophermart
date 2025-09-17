package handlers

import (
	"errors"
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/gin-gonic/gin"
)

type Login struct {
	service service.Service
}

func NewLogin(service service.Service) *Login {
	return &Login{service: service}
}

func (l Login) Handle(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !user.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("login and password should not be empty").Error()})
		return
	}
	token, err := l.service.LoginUser(c, user)
	if errors.Is(err, serviceErrors.ErrBadUserPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("Authorization", token, 3600, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "User logged in"})

}
