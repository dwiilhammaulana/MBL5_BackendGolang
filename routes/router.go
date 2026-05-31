package routes

import (
	"github.com/dwiilhammaulana/gin-firebase-backend/handlers"
	"github.com/dwiilhammaulana/gin-firebase-backend/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Init handlers
	authHandler := handlers.NewAuthHandler()
	productHandler := handlers.NewProductHandler()
	cartHandler := handlers.NewCartHandler()
	orderHandler := handlers.NewOrderHandler()

	// API v1 group
	v1 := r.Group("/v1")
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": "gin-firebase-backend",
			})
		})

		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/verify-token", authHandler.VerifyToken)
		}

		// Protected routes (butuh Backend JWT)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			cart := protected.Group("/cart")
			{
				cart.GET("", cartHandler.GetCart)
				cart.POST("", cartHandler.AddToCart)
				cart.PUT("/:id", cartHandler.UpdateItem)
				cart.DELETE("/:id", cartHandler.RemoveItem)
				cart.DELETE("", cartHandler.ClearCart)
			}

			orders := protected.Group("/orders")
			{
				orders.POST("/checkout", orderHandler.Checkout)
				orders.GET("", orderHandler.GetMyOrders)
				orders.GET("/:id", orderHandler.GetOrderDetail)
			}

			products := protected.Group("/products")
			{
				products.GET("", productHandler.GetAll)
				products.GET("/:id", productHandler.GetByID)

				// Hanya admin
				adminProducts := products.Group("")
				adminProducts.Use(middleware.AdminOnly())
				{
					adminProducts.POST("", productHandler.Create)
					adminProducts.PUT("/:id", productHandler.Update)
					adminProducts.DELETE("/:id", productHandler.Delete)
				}
			}
		}
	}

	return r
}
