package server

import (
	"github.com/ahhcash/ghastlydb/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

type Server struct {
	db     *db.DB
	router *echo.Echo
}

func (s *Server) Router() *echo.Echo {
	return s.router
}

func NewServer(database *db.DB) *Server {
	s := &Server{
		db:     database,
		router: echo.New(),
	}

	s.router.Use(middleware.Logger())
	s.router.Use(middleware.Recover())
	s.router.Use(middleware.CORS())

	// Register routes
	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	// Health check
	s.router.GET("/health", s.handleHealthCheck)

	// Database operations
	s.router.GET("/", s.handleHealthCheck)
	s.router.POST("/v1/documents", s.handlePut)
	s.router.GET("/v1/documents/:key", s.handleGet)
	s.router.DELETE("/v1/documents/:key", s.handleDelete)
	s.router.POST("/v1/search", s.handleSearch)
	s.router.GET("/v1/config", s.handleGetConfig)
}

// Start begins listening for HTTP requests
func (s *Server) Start(port string) error {
	return s.router.Start(port)
}

// handleHealthCheck responds to health check requests
func (s *Server) handleHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// PutRequest represents the request body for storing documents
type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// handlePut handles document storage requests
func (s *Server) handlePut(c echo.Context) error {
	var req PutRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := s.db.Put(req.Key, req.Value); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"status": "success",
	})
}

// handleGet retrieves documents by key
func (s *Server) handleGet(c echo.Context) error {
	key := c.Param("key")
	value, err := s.db.Get(key)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Document not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"key":   key,
		"value": value,
	})
}

// handleDelete removes documents
func (s *Server) handleDelete(c echo.Context) error {
	key := c.Param("key")
	if err := s.db.Delete(key); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Document not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}

// SearchRequest represents the search query parameters
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// handleSearch performs semantic search over documents
func (s *Server) handleSearch(c echo.Context) error {
	var req SearchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	results, err := s.db.Search(req.Query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"results": results,
	})
}

// handleGetConfig returns the current database configuration
func (s *Server) handleGetConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.DBConfig)
}
