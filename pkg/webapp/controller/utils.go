package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseIDParam(ctx *gin.Context) (int, error) {
	return parseIntParam(ctx, "id")
}

func parseIntParam(ctx *gin.Context, name string) (int, error) {
	id, err := strconv.Atoi(ctx.Param(name))
	return id, err
}
