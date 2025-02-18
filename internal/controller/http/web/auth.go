package web

import (
	"gitea.chrnv.ru/vanadium23/kompanion/internal/auth"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/logger"
	"github.com/gin-gonic/gin"
)

type authRoutes struct {
	auth auth.AuthInterface
	l    logger.Interface
}

func newAuthRoutes(handler *gin.RouterGroup, a auth.AuthInterface, l logger.Interface) {
	r := &authRoutes{a, l}

	handler.GET("/login", r.loginForm)
	handler.POST("/login", r.loginAction)
	handler.GET("/logout", r.logoutAction)
}

func (r *authRoutes) loginForm(c *gin.Context) {
	c.HTML(200, "login", passStandartContext(c, gin.H{}))
}

func (r *authRoutes) logoutAction(c *gin.Context) {
	sessionKey, err := c.Cookie("session")
	if err != nil {
		c.Redirect(302, "/auth/login")
		return
	}
	r.auth.Logout(c.Request.Context(), sessionKey)
	c.SetCookie("session", "", 0, "/", "", false, true)
	c.Redirect(302, "/auth/login")
}

func (r *authRoutes) loginAction(c *gin.Context) {
	clientIP, _ := c.RemoteIP()
	sessionKey, err := r.auth.Login(
		c.Request.Context(),
		c.PostForm("username"),
		c.PostForm("password"),
		c.Request.UserAgent(),
		clientIP,
	)
	if err != nil {
		r.l.Error(err)
		c.HTML(200, "login", passStandartContext(c, gin.H{"error": err.Error()}))
		return
	}
	c.SetCookie("session", sessionKey, 0, "/", "", false, true)
	c.Redirect(302, "/books")
}

func authMiddleware(a auth.AuthInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey, err := c.Cookie("session")
		if err != nil {
			c.Redirect(302, "/auth/login")
			c.Abort()
			return
		}

		if !a.IsAuthenticated(c.Request.Context(), sessionKey) {
			c.Redirect(302, "/auth/login")
			c.Abort()
			return
		}
		c.Set("isAuthenticated", true)
		c.Next()
	}
}
