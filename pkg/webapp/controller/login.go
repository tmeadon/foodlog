package controller

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/tmeadon/foodlog/pkg/data"
	"github.com/tmeadon/foodlog/pkg/webapp/middleware"
)

func (c *Controller) loadLoginRoutes() {
	login := c.publicRoutes.Group("/login")
	login.GET("/", c.loginGet)
	login.POST("/", c.loginPost)

	c.publicRoutes.GET("/logout", c.logout)
}

func (c *Controller) loginGet(ctx *gin.Context) {
	session := sessions.Default(ctx)
	user := session.Get(middleware.UserKey)
	if user != nil {
		ctx.HTML(http.StatusBadRequest, "views/login.html",
			gin.H{
				"content": "Please logout first",
				"user":    user,
			})
		return
	}
	ctx.HTML(http.StatusOK, "views/login.html", gin.H{
		"content": "",
		"user":    user,
	})
}

func (c *Controller) loginPost(ctx *gin.Context) {
	session := sessions.Default(ctx)
	user := session.Get(middleware.UserKey)
	if user != nil {
		ctx.HTML(http.StatusBadRequest, "views/login.html", gin.H{"content": "Please logout first"})
		return
	}

	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
		ctx.HTML(http.StatusBadRequest, "views/login.html", gin.H{"content": "Parameters can't be empty"})
		return
	}

	valid := c.validateUser(username, password)
	if !valid {
		log.Printf("failed login for username %v from IP %v", username, ctx.Request.Header.Get("X-Forwarded-For"))
		ctx.HTML(http.StatusUnauthorized, "views/login.html", gin.H{"content": "Incorrect username or password"})
		return
	}

	log.Printf("login successful for username %v from IP %v", username, ctx.Request.Header.Get("X-Forwarded-For"))
	session.Set(middleware.UserKey, username)
	session.Options(sessions.Options{MaxAge: 24 * 60 * 60, Path: "/"})
	if err := session.Save(); err != nil {
		log.Printf("failed to save session: %v", err)
		ctx.HTML(http.StatusInternalServerError, "views/login.html", gin.H{"content": "Failed to save session"})
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, "/")
}

func (c *Controller) validateUser(username string, password string) bool {
	user, err := data.GetUserByUsernameWithPasswordHash(username)

	if err != nil {
		log.Printf("failed to find user '%v' in database: %v", username, err)
		return false
	}

	valid, err := user.ValidatePassword(password)
	if err != nil {
		log.Printf("failed to validate password: %v", err)
	}

	return valid
}

func (c *Controller) logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Delete(middleware.UserKey)

	if err := session.Save(); err != nil {
		ctx.Redirect(http.StatusFound, "/")
		log.Println("failed to save session:", err)
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}
