package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	// config := config.LoadConfig()

	// Initialize repositories
	// userRepo := repository.NewUserRepository(db)
	// productRepo := repository.NewProductRepository(db)
	// orderRepo := repository.NewOrderRepository(db)
	// inventoryRepo := repository.NewInventoryRepository(db)

	// Initialize services
	// userService := service.NewUserService(userRepo)
	// productService := service.NewProductService(productRepo)
	// orderService := service.NewOrderService(orderRepo, productRepo, inventoryRepo)

	// Initialize handlers
	// userHandler := api.NewUserHandler(userService)
	// productHandler := api.NewProductHandler(productService)
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
			authRoutes.POST("/register", stubRegisterHandler)
			authRoutes.POST("/login", stubLoginHandler)
		}

		// Product routes
		productRoutes := apiRoutes.Group("/products")
		{
			productRoutes.GET("", stubGetProductsHandler)
			productRoutes.GET("/:id", stubGetProductHandler)
			productRoutes.GET("/merchant/:id", stubGetProductsByMerchantHandler)
			productRoutes.POST("", stubCreateProductHandler)
			productRoutes.PUT("/:id", stubUpdateProductHandler)
			productRoutes.DELETE("/:id", stubDeleteProductHandler)
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

// Stub handlers for testing
func stubRegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Return a mock user
	c.JSON(201, gin.H{
		"id":       1,
		"username": req.Username,
		"email":    req.Email,
		"role":     req.Role,
	})
}

func stubLoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Return a mock user and token
	c.JSON(200, gin.H{
		"user": gin.H{
			"id":       1,
			"username": "demoUser",
			"email":    req.Email,
			"role":     "merchant",
		},
		"token": "mock-jwt-token-for-testing",
	})
}

func stubGetProductsHandler(c *gin.Context) {
	// Return mock products
	c.JSON(200, []gin.H{
		{
			"id":           1,
			"merchant_id":  1,
			"name":         "Mojito",
			"description":  "Classic cocktail with rum, mint, and lime",
			"price":        8.99,
			"category":     "Cocktails",
			"image_url":    "https://via.placeholder.com/300x200.png?text=Mojito",
			"is_available": true,
		},
		{
			"id":           2,
			"merchant_id":  1,
			"name":         "Old Fashioned",
			"description":  "Whiskey cocktail with sugar and bitters",
			"price":        9.99,
			"category":     "Cocktails",
			"image_url":    "https://via.placeholder.com/300x200.png?text=Old+Fashioned",
			"is_available": true,
		},
		{
			"id":           3,
			"merchant_id":  2,
			"name":         "Margarita",
			"description":  "Tequila cocktail with lime and salt",
			"price":        7.99,
			"category":     "Cocktails",
			"image_url":    "https://via.placeholder.com/300x200.png?text=Margarita",
			"is_available": true,
		},
	})
}

func stubGetProductHandler(c *gin.Context) {
	id := c.Param("id")

	// Return a mock product based on ID
	c.JSON(200, gin.H{
		"id":           id,
		"merchant_id":  1,
		"name":         "Mojito",
		"description":  "Classic cocktail with rum, mint, and lime",
		"price":        8.99,
		"category":     "Cocktails",
		"image_url":    "https://via.placeholder.com/300x200.png?text=Mojito",
		"is_available": true,
		"ingredients": []gin.H{
			{
				"id":       1,
				"name":     "White Rum",
				"quantity": 50, // in ml
				"unit":     "ml",
			},
			{
				"id":       2,
				"name":     "Fresh Mint",
				"quantity": 10, // leaves
				"unit":     "leaves",
			},
			{
				"id":       3,
				"name":     "Lime Juice",
				"quantity": 25, // in ml
				"unit":     "ml",
			},
			{
				"id":       4,
				"name":     "Sugar Syrup",
				"quantity": 15, // in ml
				"unit":     "ml",
			},
		},
	})
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

func stubGetProductsByMerchantHandler(c *gin.Context) {
	merchantID := c.Param("id")

	// Return mock products for the given merchant ID
	c.JSON(200, []gin.H{
		{
			"id":           1,
			"merchant_id":  merchantID,
			"name":         "Mojito",
			"description":  "Classic cocktail with rum, mint, and lime",
			"price":        8.99,
			"category":     "Cocktails",
			"image_url":    "https://via.placeholder.com/300x200.png?text=Mojito",
			"is_available": true,
		},
		{
			"id":           2,
			"merchant_id":  merchantID,
			"name":         "Old Fashioned",
			"description":  "Whiskey cocktail with sugar and bitters",
			"price":        9.99,
			"category":     "Cocktails",
			"image_url":    "https://via.placeholder.com/300x200.png?text=Old+Fashioned",
			"is_available": true,
		},
	})
}

func stubCreateProductHandler(c *gin.Context) {
	// Get product data from request
	var product struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		MerchantID  uint    `json:"merchant_id"`
		ImageURL    string  `json:"image_url"`
		IsAvailable bool    `json:"is_available"`
	}

	if err := c.BindJSON(&product); err != nil {
		c.JSON(400, gin.H{"error": "Invalid product data"})
		return
	}

	// In a real app, we would save this to the database
	// For demonstration, return a mock response with the created product
	c.JSON(201, gin.H{
		"id":           1,
		"name":         product.Name,
		"description":  product.Description,
		"price":        product.Price,
		"category":     product.Category,
		"merchant_id":  product.MerchantID,
		"image_url":    product.ImageURL,
		"is_available": product.IsAvailable,
		"created_at":   time.Now(),
		"updated_at":   time.Now(),
	})
}

func stubUpdateProductHandler(c *gin.Context) {
	id := c.Param("id")

	// Get product data from request
	var product struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		MerchantID  uint    `json:"merchant_id"`
		ImageURL    string  `json:"image_url"`
		IsAvailable bool    `json:"is_available"`
	}

	if err := c.BindJSON(&product); err != nil {
		c.JSON(400, gin.H{"error": "Invalid product data"})
		return
	}

	// In a real app, we would update the product in the database
	// For demonstration, return a mock response with the updated product
	c.JSON(200, gin.H{
		"id":           id,
		"name":         product.Name,
		"description":  product.Description,
		"price":        product.Price,
		"category":     product.Category,
		"merchant_id":  product.MerchantID,
		"image_url":    product.ImageURL,
		"is_available": product.IsAvailable,
		"updated_at":   time.Now(),
	})
}

func stubDeleteProductHandler(c *gin.Context) {
	id := c.Param("id")

	// In a real app, we would delete the product from the database
	// For demonstration, just return a success response
	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Product %s deleted successfully", id),
	})
}
