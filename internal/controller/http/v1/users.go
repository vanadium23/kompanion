package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/pkg/logger"
)

type userRoutes struct {
	auth auth.AuthInterface
	l    logger.Interface
}

func newUserRoutes(handler *gin.RouterGroup, a auth.AuthInterface, l logger.Interface) {
	r := &userRoutes{a, l}

	h := handler.Group("/users")
	h.Use(authDeviceMiddleware(a, l))
	{
		h.GET("/auth", r.authenicate)
	}
}

func (r *userRoutes) authenicate(c *gin.Context) {
	// authenication done by authDeviceMiddleware
	c.AsciiJSON(http.StatusOK, gin.H{"message": "OK", "code": 200})
}

func authDeviceMiddleware(auth auth.AuthInterface, l logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("x-auth-user")
		hashed_password := c.GetHeader("x-auth-key")
		if username == "" || hashed_password == "" {
			c.AsciiJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
			c.Abort()
			return
		}
		if !auth.CheckDevicePassword(c.Request.Context(), username, hashed_password, false) {
			c.AsciiJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
			c.Abort()
			return
		}
		c.Set("device_name", username)
		c.Next()
	}
}
