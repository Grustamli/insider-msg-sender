package api

import (
	"github.com/gin-gonic/gin"
	"github.com/grustamli/insider-msg-sender/application"
	"github.com/grustamli/insider-msg-sender/daemon"
	docs "github.com/grustamli/insider-msg-sender/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @Title Insider Message Sender API
// @version 1.0
// @description API endpoints for the Insider Message Sender that periodically sends messages from DB
// @contact.name Gadir Rustamli
// @contact.email gadir.rustamli@outlook.com
// @host localhost:8000
// @BasePath /
// @accept json
// @produce json
// @schemes http
// @tag.name Scheduler

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
	s.router.GET("/messages", s.listSentMessages)
}

func (s *Server) registerSwagger() {
	docs.SwaggerInfo.BasePath = "/"
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
