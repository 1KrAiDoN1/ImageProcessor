package httpserver

import (
	"context"
	"errors"
	"fmt"
	"imageprocessor/backend/internal/config"
	"imageprocessor/backend/internal/http-server/handler"
	"imageprocessor/backend/internal/http-server/middleware"
	"imageprocessor/backend/internal/http-server/routes"

	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	server   *http.Server
	router   *gin.Engine
	handlers *handler.Handler
	logger   *zap.Logger
	config   *config.ServiceConfig
}

func NewServer(logger *zap.Logger, cfg *config.ServiceConfig, handlers *handler.Handler) *Server {
	router := gin.New()

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		server:   srv,
		router:   router,
		handlers: handlers,
		logger:   logger,
		config:   cfg,
	}
}

func (s *Server) Run() error {
	s.setupRoutes()

	s.logger.Info("Starting HTTP server", zap.String("address", s.config.Server.Address))

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.logger.Info("Shutting down HTTP server...")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", zap.Error(err))
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	s.logger.Info("HTTP server gracefully shut down")
	return nil
}

func (s *Server) setupRoutes() {
	// CORS middleware для работы с frontend
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost", "http://localhost:80", "http://localhost:8000", "http://localhost:3000", "http://127.0.0.1:8000", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "image-processor-api",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := s.router.Group("/api/v1")
	api.Use(middleware.Logger(s.logger))
	routes.SetupRoutes(api, s.handlers, s.logger)
}
