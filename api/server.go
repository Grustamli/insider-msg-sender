package api

import (
	"github.com/gin-gonic/gin"
	"insider-message-sender/application"
	"insider-message-sender/daemon"
)

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
