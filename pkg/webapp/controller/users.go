package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmeadon/foodlog/pkg/data"
	"github.com/tmeadon/foodlog/pkg/webapp/middleware"
)

func (c *Controller) loadUserRoutes() {
	g := c.privateRoutes.Group("/users")

	g.GET("/", c.allUsers)
	g.POST("/new", c.createUser)
	g.POST("/:id", c.updateUser)
	g.POST("/:id/delete", c.deleteUser)
	g.POST("/:id/password", c.changeUserPassword)
}

func (c *Controller) allUsers(ctx *gin.Context) {
	users := c.getAllUsers()
	ctx.HTML(
		http.StatusOK,
		"views/users.html",
		gin.H{
			"Users":    users,
			"Username": middleware.CurrentUser(ctx).Username,
		},
	)
}

func (c *Controller) getAllUsers() []data.User {
	users, err := data.GetUsers()
	if err != nil {
		panic(err)
	}
	return users
}

func (c *Controller) getUserAndHandleErrors(id int, ctx *gin.Context) (*data.User, bool) {
	u, err := data.GetUserById(id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			ctx.AbortWithStatus(http.StatusNotFound)
			return nil, false
		}
		panic(err)
	}
	return u, true
}

func (c *Controller) createUser(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	if username == "" || password == "" {
		u := c.getAllUsers()
		ctx.HTML(http.StatusBadRequest, "views/users.html", gin.H{
			"Users":    u,
			"Error":    "Username and password can't be empty",
			"Username": middleware.CurrentUser(ctx).Username,
		})
		return
	}

	u, _ := data.NewUser(ctx.PostForm("username"), ctx.PostForm("password"))

	if err := data.SaveUser(u); err != nil {
		var e string

		if errors.Is(err, data.ErrUniqueConstraintFailed) {
			e = "username already exists"
		} else {
			log.Print(err)
			e = "unexpected error occurred"
		}

		u := c.getAllUsers()
		ctx.HTML(http.StatusBadRequest, "views/users.html", gin.H{
			"Users":    u,
			"Error":    e,
			"Username": middleware.CurrentUser(ctx).Username,
		})
		return
	}

	ctx.Redirect(http.StatusFound, "/users")
}

func (c *Controller) saveUser(u *data.User) {
	err := data.SaveUser(u)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) updateUser(ctx *gin.Context) {
	id, err := parseIDParam(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/users")
		return
	}

	u, ok := c.getUserAndHandleErrors(id, ctx)
	if !ok {
		return
	}

	c.saveUser(u)
	ctx.Redirect(http.StatusFound, "/users")
}

func (c *Controller) deleteUser(ctx *gin.Context) {
	id, err := parseIDParam(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/users")
		return
	}

	u, ok := c.getUserAndHandleErrors(id, ctx)
	if !ok {
		return
	}

	err = data.DeleteUser(u)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) changeUserPassword(ctx *gin.Context) {
	id, err := parseIDParam(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/users")
		return
	}

	u, ok := c.getUserAndHandleErrors(id, ctx)
	if !ok {
		return
	}

	newPass := ctx.PostForm("password")

	if newPass == "" {
		users := c.getAllUsers()
		ctx.HTML(http.StatusBadRequest, "views/users.html", gin.H{
			"Users":    users,
			"Error":    "Password can't be empty",
			"Username": middleware.CurrentUser(ctx).Username,
		})
	}

	u.SetPassword(newPass)
	c.saveUser(u)
	ctx.Redirect(http.StatusFound, "/users")
}
