package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/kexincchen/homebar/internal/api"
	"github.com/kexincchen/homebar/internal/config"
	"github.com/kexincchen/homebar/internal/db"
	"github.com/kexincchen/homebar/internal/repository"
	"github.com/kexincchen/homebar/internal/service"
	"io/ioutil"
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
	// orderRepo := repository.NewOrderRepository(db)
	// inventoryRepo := repository.NewInventoryRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	// orderService := service.NewOrderService(orderRepo, productRepo, inventoryRepo)

	// Initialize handlers
	userHandler := api.NewUserHandler(userService)
	productHandler := api.NewProductHandler(productService)
	// orderHandler := api.NewOrderHandler(orderService)

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
			// For now, we'll use stub handlers until we implement the full functionality
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
			orderRoutes.POST("", stubCreateOrderHandler)
			orderRoutes.GET("", stubGetOrdersHandler)
			orderRoutes.GET("/:id", stubGetOrderHandler)
		}

		// Merchant routes
		merchantRoutes := apiRoutes.Group("/merchants")
		{
			merchantRoutes.GET("", stubGetMerchantsHandler)
			merchantRoutes.GET("/:id", stubGetMerchantHandler)
			merchantRoutes.GET("/username/:username", stubGetMerchantByUsernameHandler)
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

func stubCreateOrderHandler(c *gin.Context) {
	var req struct {
		CustomerID uint `json:"customer_id"`
		MerchantID uint `json:"merchant_id"`
		Items      []struct {
			ProductID uint `json:"product_id"`
			Quantity  int  `json:"quantity"`
		} `json:"items"`
		Notes string `json:"notes"`
	}

	// Print the raw request body for debugging
	body, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	fmt.Println("Request body:", string(body))

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.CustomerID == 0 || req.MerchantID == 0 || len(req.Items) == 0 {
		c.JSON(400, gin.H{
			"error": "Missing required fields",
		})
		return
	}

	// Return a mock order
	c.JSON(201, gin.H{
		"id":           1,
		"customer_id":  req.CustomerID,
		"merchant_id":  req.MerchantID,
		"total_amount": 26.97,
		"status":       "pending",
		"notes":        req.Notes,
		"created_at":   "2023-09-20T15:04:05Z",
	})
}

func stubGetOrdersHandler(c *gin.Context) {
	// Return mock orders
	c.JSON(200, []gin.H{
		{
			"id":           1,
			"customer_id":  1,
			"merchant_id":  1,
			"total_amount": 26.97,
			"status":       "pending",
			"notes":        "No ice in mojito",
			"created_at":   "2023-09-20T15:04:05Z",
		},
		{
			"id":           2,
			"customer_id":  1,
			"merchant_id":  2,
			"total_amount": 15.98,
			"status":       "delivered",
			"notes":        "",
			"created_at":   "2023-09-19T12:34:56Z",
		},
	})
}

func stubGetOrderHandler(c *gin.Context) {
	id := c.Param("id")

	// Return a mock order based on ID
	c.JSON(200, gin.H{
		"id":           id,
		"customer_id":  1,
		"merchant_id":  1,
		"total_amount": 26.97,
		"status":       "pending",
		"notes":        "No ice in mojito",
		"created_at":   "2023-09-20T15:04:05Z",
		"items": []gin.H{
			{
				"id":           1,
				"product_id":   1,
				"product_name": "Mojito",
				"quantity":     2,
				"price":        8.99,
			},
			{
				"id":           2,
				"product_id":   2,
				"product_name": "Old Fashioned",
				"quantity":     1,
				"price":        9.99,
			},
		},
	})
}

func stubGetMerchantsHandler(c *gin.Context) {
	// Return a list of mock merchants
	c.JSON(200, []gin.H{
		{
			"id":            1,
			"user_id":       2,
			"business_name": "Cocktail Haven",
			"description":   "Specializing in premium cocktails with a modern twist",
			"address":       "123 Main St, City",
			"phone":         "555-123-4567",
			"is_verified":   true,
			"username":      "cocktailhaven",
		},
		{
			"id":            2,
			"user_id":       3,
			"business_name": "Tropical Tastes",
			"description":   "Exotic beach-inspired drinks and cocktails",
			"address":       "456 Ocean Ave, Beach City",
			"phone":         "555-987-6543",
			"is_verified":   true,
			"username":      "tropicaltastes",
		},
	})
}

func stubGetMerchantHandler(c *gin.Context) {
	id := c.Param("id")

	// Return a mock merchant based on ID
	c.JSON(200, gin.H{
		"id":            id,
		"user_id":       2,
		"business_name": "Cocktail Haven",
		"description":   "Specializing in premium cocktails with a modern twist",
		"address":       "123 Main St, City",
		"phone":         "555-123-4567",
		"is_verified":   true,
		"username":      "cocktailhaven",
	})
}

func stubGetMerchantByUsernameHandler(c *gin.Context) {
	username := c.Param("username")

	// In a real app, we would search for the merchant by username
	// For demo purposes, just return a mock merchant with the given username
	c.JSON(200, gin.H{
		"id":            1,
		"user_id":       2,
		"business_name": "Cocktail Haven",
		"description":   "Specializing in premium cocktails with a modern twist",
		"address":       "123 Main St, City",
		"phone":         "555-123-4567",
		"is_verified":   true,
		"username":      username,
	})
}
