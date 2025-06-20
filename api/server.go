package api

import (
	"github.com/gin-gonic/gin"
	"github.com/grustamli/insider-msg-sender/application"
	"github.com/grustamli/insider-msg-sender/daemon"
	docs "github.com/grustamli/insider-msg-sender/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @BasePath /

type Server struct {
	app       application.App
	scheduler daemon.Daemon
	router    *gin.Engine
	port      string
}

func NewServer(router *gin.Engine, port string, app application.App, scheduler daemon.Daemon) *Server {
	s := &Server{
		router:    router,
		app:       app,
		scheduler: scheduler,
		port:      port,
	}
	s.initHandlers()
	s.registerSwagger()
	return s
}

func (s *Server) Run() error {
	return s.router.Run(s.port)
}

func (s *Server) initHandlers() {
	s.router.POST("/start", s.startSender)
	s.router.POST("/stop", s.stopSender)
	s.router.GET("/sent-messages", s.listSentMessages)
}

func (s *Server) registerSwagger() {
	docs.SwaggerInfo.BasePath = "/"
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
