package web

import (
	"gitea.chrnv.ru/vanadium23/kompanion/internal/auth"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/logger"
	"github.com/gin-gonic/gin"
)

type deviceRoutes struct {
	auth auth.AuthInterface
	l    logger.Interface
}

func newDeviceRoutes(handler *gin.RouterGroup, a auth.AuthInterface, l logger.Interface) {
	r := &deviceRoutes{a, l}

	handler.GET("/", r.listDevices)
	handler.POST("/add", r.addDeviceAction)
	handler.POST("/deactivate/:device_name", r.deactivateDeviceAction)
}

func (r *deviceRoutes) listDevices(c *gin.Context) {
	devices, err := r.auth.ListDevices(c.Request.Context())
	if err != nil {
		c.HTML(500, "devices", passStandartContext(c, gin.H{
			"error": "Failed to load devices",
		}))
		return
	}

	c.HTML(200, "devices", passStandartContext(c, gin.H{
		"devices": devices,
	}))
}

func (r *deviceRoutes) addDeviceAction(c *gin.Context) {
	deviceName := c.PostForm("device_name")
	password := c.PostForm("password")

	if deviceName == "" || password == "" {
		c.HTML(400, "devices", passStandartContext(c, gin.H{
			"error": "Device name and password are required",
		}))
		return
	}

	err := r.auth.AddUserDevice(c.Request.Context(), deviceName, password)
	if err != nil {
		c.HTML(400, "devices", passStandartContext(c, gin.H{
			"error": err.Error(),
		}))
		return
	}

	c.Redirect(302, "/devices")
}

func (r *deviceRoutes) deactivateDeviceAction(c *gin.Context) {
	deviceName := c.Param("device_name")
	err := r.auth.DeactivateUserDevice(c.Request.Context(), deviceName)
	if err != nil {
		c.HTML(400, "devices", passStandartContext(c, gin.H{
			"error": err.Error(),
		}))
		return
	}

	c.Redirect(302, "/devices")
}
