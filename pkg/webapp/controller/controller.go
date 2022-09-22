package controller

import (
	"github.com/gin-gonic/gin"
)

type Controller struct {
	publicRoutes  *gin.RouterGroup
	privateRoutes *gin.RouterGroup
}

func NewController(publicRoutes *gin.RouterGroup, privateRoutes *gin.RouterGroup) *Controller {
	c := Controller{
		publicRoutes:  publicRoutes,
		privateRoutes: privateRoutes,
	}

	c.loadRoutes()
	return &c
}

func (c *Controller) loadRoutes() {
	c.loadLoginRoutes()
	c.loadHomepageRoutes()
	c.loadUserRoutes()
}
