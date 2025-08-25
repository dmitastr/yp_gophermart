package handlers

import (
	"net/http"

	"github.com/dmitastr/yp_gophermart/internal/domain/models"
	"github.com/dmitastr/yp_gophermart/internal/domain/service"
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
	token, err := r.service.RegisterUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("Authorization", token, 3600, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully!"})
}
