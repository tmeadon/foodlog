package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmeadon/foodlog/pkg/data"
	"github.com/tmeadon/foodlog/pkg/webapp/middleware"
)

func (c *Controller) loadHomepageRoutes() {
	c.privateRoutes.GET("/", c.homepage)
	c.privateRoutes.POST("/entry", c.newEntry)
	c.privateRoutes.DELETE("/entry/:id", c.deleteEntry)
	c.privateRoutes.POST("/entry/:id", c.editEntry)
}

func (c *Controller) homepage(ctx *gin.Context) {
	user := middleware.CurrentUser(ctx)
	entries, err := data.GetEntriesByUser(user.Id)

	if err != nil {
		log.Printf("failed loading homepage for user %v: %v", user.Id, err)
		ctx.HTML(http.StatusInternalServerError, "views/homepage.html", gin.H{"Username": user.Username, "Error": "Loading failed"})
		return
	}

	ctx.HTML(
		http.StatusOK,
		"views/homepage.html",
		gin.H{
			"Username": user.Username,
			"Entries":  entries,
		},
	)
}

func (c *Controller) newEntry(ctx *gin.Context) {
	user := middleware.CurrentUser(ctx)

	eattime, err := time.Parse("2006-01-02T15:04", ctx.PostForm("time"))
	if err != nil {
		log.Printf("failed parsing time %v: %v", ctx.PostForm("time"), err)
		ctx.HTML(http.StatusBadRequest, "views/homepage.html", gin.H{"Username": user.Username, "Error": "Invalid time submitted"})
		return
	}

	entry := data.LogEntry{
		UserId: user.Id,
		Time:   eattime,
		Food:   ctx.PostForm("food"),
		Notes:  ctx.PostForm("notes"),
	}

	err = data.SaveEntry(&entry)
	if err != nil {
		log.Printf("failed saving entry %#v: %v", entry, err)
		ctx.HTML(http.StatusBadRequest, "views/homepage.html", gin.H{"Username": user.Username, "Error": "Something went wrong"})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (c *Controller) deleteEntry(ctx *gin.Context) {
	user := middleware.CurrentUser(ctx)

	entryId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("failed converting id param: %v", err)
		ctx.Redirect(http.StatusFound, "/")
	}

	entry, err := data.GetEntryById(entryId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			log.Printf("entry with id %v not found in database", entryId)
			ctx.Redirect(http.StatusFound, "/")
		} else {
			log.Printf("failed to find entry with id %v in database: %v", entryId, err)
			ctx.Redirect(http.StatusFound, "/")
		}
		return
	}

	if entry.UserId != user.Id {
		log.Printf("entry with id %v does not belong to current user %v", entryId, user.Username)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	err = data.DeleteEntry(entry)
	if err != nil {
		log.Printf("failed to delete entry %v: %v", entry.Id, err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}
}

func (c *Controller) editEntry(ctx *gin.Context) {
	user := middleware.CurrentUser(ctx)

	entryId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("failed converting id param: %v", err)
		ctx.Redirect(http.StatusFound, "/")
	}

	entry, err := data.GetEntryById(entryId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			log.Printf("entry with id %v not found in database", entryId)
			ctx.Redirect(http.StatusFound, "/")
		} else {
			log.Printf("failed to find entry with id %v in database: %v", entryId, err)
			ctx.Redirect(http.StatusFound, "/")
		}
		return
	}

	if entry.UserId != user.Id {
		log.Printf("entry with id %v does not belong to current user %v", entryId, user.Username)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	eattime, err := time.Parse("2006-01-02T15:04", ctx.PostForm("time"))
	if err != nil {
		log.Printf("failed parsing time %v: %v", ctx.PostForm("time"), err)
		ctx.HTML(http.StatusBadRequest, "views/homepage.html", gin.H{"Username": user.Username, "Error": "Invalid time submitted"})
		return
	}

	entry.Time = eattime
	entry.Food = ctx.PostForm("food")

	err = data.SaveEntry(entry)
	if err != nil {
		log.Printf("failed saving entry %#v: %v", entry, err)
		ctx.HTML(http.StatusBadRequest, "views/homepage.html", gin.H{"Username": user.Username, "Error": "Something went wrong"})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}
