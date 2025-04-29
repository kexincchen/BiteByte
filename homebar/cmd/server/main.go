package main

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kexincchen/homebar/internal/api"
	"github.com/kexincchen/homebar/internal/config"
	"github.com/kexincchen/homebar/internal/db"
	"github.com/kexincchen/homebar/internal/repository"
	"github.com/kexincchen/homebar/internal/repository/postgres"
	"github.com/kexincchen/homebar/internal/service"
	//"time"
)

func main() {
	// Initialize configuration
	cfg := config.Load()
	dbConn, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("db init error: %v", err)
	}
	defer func(dbConn *sql.DB) {
		err := dbConn.Close()
		if err != nil {
			log.Fatalf("db close error: %v", err)
		}
	}(dbConn)

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbConn)
	customerRepo := repository.NewCustomerRepository(dbConn)
	productRepo := repository.NewProductRepository(dbConn)
	orderRepo := repository.NewOrderRepository(dbConn)
	merchantRepo := repository.NewMerchantRepository(dbConn)
	ingredientRepo := postgres.NewIngredientRepository(dbConn)
	productIngredientRepo := postgres.NewProductIngredientRepository(dbConn)
	inventoryRepo := postgres.NewInventoryRepository(dbConn)

	// Initialize services with all repositories
	userService := service.NewUserService(userRepo, customerRepo, merchantRepo, dbConn)
	productService := service.NewProductService(productRepo)
	ingredientService := service.NewIngredientService(
		ingredientRepo,
		productIngredientRepo,
	)
	productIngredientService := service.NewProductIngredientService(productIngredientRepo)
	orderService := service.NewOrderService(
		orderRepo,
		productRepo,
		ingredientService,
	)
	merchantService := service.NewMerchantService(merchantRepo)

	// Initialize handlers
	userHandler := api.NewUserHandler(userService)
	productHandler := api.NewProductHandler(productService, ingredientService)
	orderHandler := api.NewOrderHandler(orderService)
	merchantHandler := api.NewMerchantHandler(merchantService)
	ingredientHandler := api.NewIngredientHandler(ingredientService)
	productIngredientHandler := api.NewProductIngredientHandler(
		productIngredientService,
		productService,
		ingredientService,
	)

	// Setup router
	router := gin.Default()

	// Enable CORS middleware
	router.Use(corsMiddleware())

	// Add logger middleware
	router.Use(loggerMiddleware())

	// Define routes
	apiRoutes := router.Group("/api")
	{
		// Auth routes
		authRoutes := apiRoutes.Group("/auth")
		{
			authRoutes.POST("/register", userHandler.Register)
			authRoutes.POST("/login", userHandler.Login)
		}

		// Product routes
		productRoutes := apiRoutes.Group("/products")
		{
			productRoutes.POST("", productHandler.Create)
			productRoutes.GET("/:id", productHandler.GetByID)
			productRoutes.PUT("/:id", productHandler.Update)
			productRoutes.DELETE("/:id", productHandler.Delete)
			productRoutes.GET("/merchant/:id", productHandler.GetByMerchant)
			productRoutes.GET("", productHandler.GetAll)
			productRoutes.GET("/:id/image", productHandler.GetImage)
			productRoutes.POST("/availability", productHandler.CheckAvailability)
		}

		// Order routes
		orderRoutes := apiRoutes.Group("/orders")
		{
			orderRoutes.POST("", orderHandler.Create)
			orderRoutes.GET("", orderHandler.List)
			orderRoutes.GET("/:id", orderHandler.GetByID)
			orderRoutes.PUT("/:id/status", orderHandler.UpdateStatus)
			orderRoutes.PUT("/:id", orderHandler.UpdateOrder)
		}

		// Merchant routes
		merchantRoutes := apiRoutes.Group("/merchants")
		{
			merchantRoutes.POST("", merchantHandler.Create)
			merchantRoutes.GET("", merchantHandler.List)
			merchantRoutes.GET("/:id", merchantHandler.GetByID)
			merchantRoutes.GET("/username/:username", merchantHandler.GetByUsername)
			merchantRoutes.GET("/user/:userID", merchantHandler.GetByUserID)
		}

		// Ingredient routes
		ingredientRoutes := apiRoutes.Group("/merchants/:id/inventory")
		{
			ingredientRoutes.GET("", ingredientHandler.GetAll)
			ingredientRoutes.POST("", ingredientHandler.Create)
			ingredientRoutes.GET("/summary", ingredientHandler.GetInventorySummary)
			ingredientRoutes.GET("/:ingredientId", ingredientHandler.GetByID)
			ingredientRoutes.PUT("/:ingredientId", ingredientHandler.Update)
			ingredientRoutes.DELETE("/:ingredientId", ingredientHandler.Delete)
		}

		// Product ingredient routes
		productIngredientRoutes := apiRoutes.Group("/products/:id/ingredients")
		{
			productIngredientRoutes.GET("", productIngredientHandler.GetByProductID)
			productIngredientRoutes.POST("", productIngredientHandler.Create)
			productIngredientRoutes.PUT("/:ingredientId", productIngredientHandler.Update)
			productIngredientRoutes.DELETE("/:ingredientId", productIngredientHandler.Delete)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// CORS middleware to allow frontend to access the API
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger middleware
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request details
		startTime := time.Now()
		requestID := uuid.New().String()

		// Set request ID header for tracking
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Get request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		log.Printf("[%s] API Request: %s %s\nHeaders: %v\nBody: %s",
			requestID, c.Request.Method, c.Request.URL.Path,
			c.Request.Header, string(requestBody))

		// Use ResponseWriter wrapper to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Log response
		latency := time.Since(startTime)
		log.Printf("[%s] API Response: %d %s (%s)\nBody: %s",
			requestID, c.Writer.Status(), http.StatusText(c.Writer.Status()),
			latency, blw.body.String())
	}
}

// Response body logger
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
