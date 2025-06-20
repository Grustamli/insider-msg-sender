// Package api defines the HTTP API server for the Insider Message Sender service.
// It registers endpoints to control scheduling and inspect sent messages,
// configures middleware, and serves Swagger documentation.
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/grustamli/insider-msg-sender/application"
	"github.com/grustamli/insider-msg-sender/daemon"
	docs "github.com/grustamli/insider-msg-sender/docs"
	"github.com/rs/zerolog"
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

// Server orchestrates the Gin router, application logic, and scheduler daemon.
// It exposes HTTP endpoints to start/stop message scheduling and to list sent messages.
type Server struct {
	app       application.App // core application business logic
	scheduler daemon.Daemon   // background scheduler for sending messages
	router    *gin.Engine     // Gin HTTP router
	port      string          // address and port for the server to bind
	log       zerolog.Logger  // structured logger for request-level logging
}

// NewServer constructs a new API server with the provided Gin engine, listening port,
// application logic, scheduler, and logger. It registers middleware, handlers, and Swagger docs.
func NewServer(router *gin.Engine, port string, app application.App, scheduler daemon.Daemon, log zerolog.Logger) *Server {
	s := &Server{
		router:    router,
		app:       app,
		scheduler: scheduler,
		port:      port,
		log:       log,
	}
	s.initMiddleware()
	s.initHandlers()
	s.registerSwagger()
	return s
}

// Run starts the HTTP server on the configured port.
// It blocks until the server exits or an error occurs.
func (s *Server) Run() error {
	return s.router.Run(s.port)
}

// initMiddleware installs global Gin middleware: request ID injection, logging, and panic recovery.
func (s *Server) initMiddleware() {
	s.router.Use(
		RequestID(),
		Logger(s.log),
		gin.Recovery(),
	)
}

// initHandlers registers HTTP routes for controlling and querying the scheduler.
// - POST /start: invoke the scheduler to begin sending messages
// - POST /stop: signal the scheduler to halt sending
// - GET /messages: return a list of all sent messages
func (s *Server) initHandlers() {
	s.router.POST("/start", s.startSender)
	s.router.POST("/stop", s.stopSender)
	s.router.GET("/messages", s.listSentMessages)
}

// registerSwagger configures the Gin route to serve Swagger UI at /swagger/*any.
func (s *Server) registerSwagger() {
	docs.SwaggerInfo.BasePath = "/"
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
