package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/tmeadon/foodlog/pkg/data"

	"net/http"
)

const UserKey string = "user"

func AuthRequired(ctx *gin.Context) {
	if IsLoggedIn(ctx) {
		ctx.Next()
		return
	}
	ctx.Redirect(http.StatusFound, "/login")
	ctx.Abort()
}

func IsLoggedIn(ctx *gin.Context) bool {
	user := loggedInUsername(ctx)
	return user != ""
}

func loggedInUsername(ctx *gin.Context) string {
	session := sessions.Default(ctx)
	usersess := session.Get(UserKey)
	username, ok := usersess.(string)

	if ok {
		return username
	}

	return ""
}

func CurrentUser(ctx *gin.Context) *data.User {
	username := loggedInUsername(ctx)

	if username == "" {
		return nil
	}

	user, err := data.GetUserByUsername(username)
	if err != nil {
		return nil
	}

	return user
}
