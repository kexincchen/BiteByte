package main

import (
	"database/sql"
	"github.com/kexincchen/homebar/internal/api"
	"github.com/kexincchen/homebar/internal/config"
	"github.com/kexincchen/homebar/internal/db"
	"github.com/kexincchen/homebar/internal/repository"
	"github.com/kexincchen/homebar/internal/service"
	"log"
	//"time"

	"github.com/gin-gonic/gin"
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
	productRepo := repository.NewProductRepository(dbConn)
	orderRepo := repository.NewOrderRepository(dbConn)
	merchantRepo := repository.NewMerchantRepository(dbConn)
	// inventoryRepo := repository.NewInventoryRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, productRepo, nil)
	merchantService := service.NewMerchantService(merchantRepo)

	// Initialize handlers
	userHandler := api.NewUserHandler(userService)
	productHandler := api.NewProductHandler(productService)
	orderHandler := api.NewOrderHandler(orderService)
	merchantHandler := api.NewMerchantHandler(merchantService)

	// Setup router
	router := gin.Default()

	// Enable CORS middleware
	router.Use(corsMiddleware())

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
		}

		// Order routes
		orderRoutes := apiRoutes.Group("/orders")
		{
			orderRoutes.POST("", orderHandler.Create)
			orderRoutes.GET("", orderHandler.List)
			orderRoutes.GET("/:id", orderHandler.GetByID)
		}

		// Merchant routes
		merchantRoutes := apiRoutes.Group("/merchants")
		{
			merchantRoutes.POST("", merchantHandler.Create)
			merchantRoutes.GET("", merchantHandler.List)
			merchantRoutes.GET("/:id", merchantHandler.GetByID)
			merchantRoutes.GET("/username/:username", merchantHandler.GetByUsername)
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
