package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kexincchen/homebar/internal/api"
	"github.com/kexincchen/homebar/internal/config"
	"github.com/kexincchen/homebar/internal/db"
	"github.com/kexincchen/homebar/internal/raft"
	"github.com/kexincchen/homebar/internal/repository"
	"github.com/kexincchen/homebar/internal/repository/postgres"
	"github.com/kexincchen/homebar/internal/service"
	//"time"
)

var (
	raftNodePtr *raft.RaftNode
	appNodeID   string
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

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
		inventoryRepo,
	)
	productIngredientService := service.NewProductIngredientService(productIngredientRepo)
	orderService := service.NewOrderService(
		orderRepo,
		productRepo,
		ingredientService,
		inventoryRepo,
	)
	merchantService := service.NewMerchantService(merchantRepo)

	// Initialize handlers
	userHandler := api.NewUserHandler(userService)
	productHandler := api.NewProductHandler(productService, ingredientService)
	orderHandler := api.NewOrderHandler(orderService, productService)
	merchantHandler := api.NewMerchantHandler(merchantService)
	ingredientHandler := api.NewIngredientHandler(ingredientService)
	productIngredientHandler := api.NewProductIngredientHandler(
		productIngredientService,
		productService,
		ingredientService,
	)

	// Initialize and start Raft BEFORE starting the HTTP server
	// Configure Raft
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "1" // Default node ID if not specified
	}

	var peerIDs []string

	if ids := os.Getenv("RAFT_PEER_IDS"); ids != "" {
		for _, id := range strings.Split(ids, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				peerIDs = append(peerIDs, id)
			}
		}
	}

	if len(peerIDs) == 0 {
		if env := os.Getenv("RAFT_PEERS"); env != "" {
			for _, kv := range strings.Split(env, ",") {
				if p := strings.SplitN(kv, "=", 2); len(p) == 2 {
					peerIDs = append(peerIDs, strings.TrimSpace(p[0]))
				}
			}
		}
	}

	if len(peerIDs) == 0 {
		peerIDs = []string{nodeID}
	}

	peerMap := map[string]string{}
	if env := os.Getenv("RAFT_PEERS"); env != "" {
		for _, kv := range strings.Split(env, ",") {
			p := strings.SplitN(kv, "=", 2)
			if len(p) == 2 {
				peerMap[p[0]] = p[1]
			}
		}
	}

	// Create Raft-enabled order service
	raftOrderService, err := service.NewRaftOrderService(
		orderService,
		nodeID,
		peerIDs,
		peerMap,
	)

	raftNode := raftOrderService.GetRaftNode()
	raftNodePtr = raftNode
	appNodeID = nodeID

	router.Use(redirectIfFollower(raftNode))
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

	rpcSrv := raft.SetupRaftRPCServer(raftNode)
	go func() {
		if err := rpcSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Raft RPC listen error: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("Failed to create Raft order service: %v", err)
	}

	// Start the Raft node
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := raftOrderService.Start(ctx); err != nil {
		log.Fatalf("Failed to start Raft node: %v", err)
	}

	// Use raftOrderService instead of orderService when initializing handlers
	orderHandler = api.NewOrderHandler(raftOrderService, productService)

	// Create and start the cluster coordinator
	raftLogger := log.New(os.Stdout, "[RAFT-COORDINATOR] ", log.LstdFlags)
	clusterCoordinator := raft.NewClusterCoordinator(raftLogger)

	// Register the node with the coordinator
	clusterCoordinator.RegisterNode(raftOrderService.GetRaftNode())

	// Start the coordinator
	if err := clusterCoordinator.Start(ctx, nodeID); err != nil {
		log.Fatalf("Failed to start cluster coordinator: %v", err)
	}
	defer clusterCoordinator.Stop()

	// Start the HTTP server (this will block)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		log.Printf("Serving on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	cancel()
}

// redirectIfFollower returns gin middleware that forwards non-leader nodes.
func redirectIfFollower(node *raft.RaftNode) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/health") ||
			strings.HasPrefix(c.Request.URL.Path, "/raft") {
			c.Next()
			return
		}

		if node.IsLeader() {
			log.Printf("[LEADER %s] handle %s %s",
				node.ID(), c.Request.Method, c.Request.URL.Path)
			c.Next()
			return
		}

		leader := node.LeaderID()
		if leader == "" {
			log.Printf("[FOLLOWER %s] leader unknown -> 503  (%s %s)",
				node.ID(), c.Request.Method, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable,
				gin.H{"error": "leader unknown"})
			return
		}

		target := fmt.Sprintf("http://localhost:90%s%s", leader, c.Request.URL.Path)
		if q := c.Request.URL.RawQuery; q != "" {
			target += "?" + q
		}

		log.Printf("[FORWARD %s] %s %s  --> leader %s  (%s)",
			node.ID(), c.Request.Method, c.Request.URL.Path, leader, target)

		c.Header("Location", target)
		c.AbortWithStatus(http.StatusTemporaryRedirect)
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
		role := "FOLLOWER"
		if raftNodePtr != nil && raftNodePtr.IsLeader() {
			role = "LEADER"
		}
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

		log.Printf("[%s][%s-%s] --> %s %s  body=%s",
			requestID, role, appNodeID,
			c.Request.Method, c.Request.URL.String(), string(requestBody))

		// Use ResponseWriter wrapper to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Log response
		latency := time.Since(startTime)
		log.Printf("[%s][%s-%s] <-- %d %s  (%s)",
			requestID, role, appNodeID,
			c.Writer.Status(), http.StatusText(c.Writer.Status()),
			latency)
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
