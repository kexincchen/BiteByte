package main

import (
	"log"

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

	// Define routes
	// api := router.Group("/api")
	// {
	//     auth := api.Group("/auth")
	//     {
	//         auth.POST("/register", userHandler.Register)
	//         auth.POST("/login", userHandler.Login)
	//     }
	//
	//     users := api.Group("/users")
	//     // users.Use(middleware.Auth()) // Add authentication middleware
	//     {
	//         users.GET("/me", userHandler.GetProfile)
	//     }
	//
	//     products := api.Group("/products")
	//     {
	//         products.GET("", productHandler.ListProducts)
	//         products.GET("/:id", productHandler.GetProduct)
	//     }
	//
	//     // More routes...
	// }

	// For now, just add a simple health check endpoint
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
