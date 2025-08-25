package middleware

import (
	"net/http"
	"strings"

	"github.com/dmitastr/yp_gophermart/internal/domain/service"
	"github.com/gin-gonic/gin"
)

type AuthorizeCheck struct {
	service service.Service
}

func NewAuthorizeCheck(service service.Service) *AuthorizeCheck {
	return &AuthorizeCheck{service: service}
}

func (a AuthorizeCheck) Handle(c *gin.Context) {
	cookie, err := c.Cookie("Authorization")
	if cookie == "" || err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization cookie"})
		return
	}
	tok := strings.TrimSpace(cookie)
	if tok == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "empty token"})
		return
	}
	claims, err := a.service.VerifyJWT(tok)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	username, _ := claims.GetSubject()
	issuer, _ := claims.GetIssuer()

	c.Set("username", username)
	c.Set("issuer", issuer)
	c.Next()

}
