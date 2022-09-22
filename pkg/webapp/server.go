package webapp

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/tmeadon/foodlog/pkg/data"
	"github.com/tmeadon/foodlog/pkg/webapp/controller"
	"github.com/tmeadon/foodlog/pkg/webapp/middleware"
)

type Server struct {
	gin        *gin.Engine
	controller *controller.Controller
}

func NewServer(dbPath string, cookieSecret []byte) *Server {
	data.Setup(dbPath)
	g := initGin(cookieSecret)
	publicGroup := g.Group("/")
	privateGroup := g.Group("/")
	privateGroup.Use(middleware.AuthRequired)

	c := controller.NewController(publicGroup, privateGroup)

	return &Server{
		gin:        g,
		controller: c,
	}
}

func initGin(cookieSecret []byte) (g *gin.Engine) {
	g = gin.New()
	g.Use(gin.Logger())
	g.Use(gin.Recovery())
	g.Use(sessions.Sessions("session", cookie.NewStore(cookieSecret)))
	g.SetTrustedProxies(nil)
	return
}

func (s *Server) Start() error {
	s.gin.Use(gin.Logger())
	s.gin.Use(gin.Recovery())

	s.gin.SetFuncMap(controller.TemplateFuncs())

	s.gin.LoadHTMLGlob("web/templates/**/*")
	s.gin.Static("/css", "./web/static/css")
	s.gin.Static("/img", "./web/static/img")
	s.gin.Static("/scss", "./web/static/scss")
	s.gin.Static("/vendor", "./web/static/vendor")
	s.gin.Static("/js", "./web/static/js")
	s.gin.StaticFile("/favicon.ico", "./web/img/favicon.ico")

	s.gin.Run()
	return nil
}
