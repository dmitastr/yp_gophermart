package handlers

import (
	"errors"
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	serviceErrors "github.com/dmitastr/yp_gophermart/internal/errors"
	"github.com/dmitastr/yp_gophermart/internal/logger"
	"github.com/gin-gonic/gin"
)

type Register struct {
	service service.Service
}

func NewRegister(service service.Service) *Register {
	return &Register{service: service}
}

func (r Register) Handle(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := r.service.RegisterUser(c, user)
	if errors.Is(err, serviceErrors.ErrUserExists) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		logger.Errorf("error while registering user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("Authorization", token, 3600, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully!"})
}
