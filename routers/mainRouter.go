package routers

import (
	"Sancai/controllers"
	"Sancai/middlewares"
	"database/sql"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "Sancai/docs"
)

func StartServer(db *sql.DB) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/login", func(c *gin.Context) { controllers.Login(c, db) })
		api.POST("/register", func(c *gin.Context) { controllers.Register(c, db) })
		api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api.Use(middlewares.JWTMiddleware())
	{
		// Customer routes
		customerRoutes := api.Group("/customers")
		{
			customerRoutes.GET("", func(c *gin.Context) { controllers.GetCustomers(c, db) })
			customerRoutes.GET("/:id", func(c *gin.Context) { controllers.GetCustomerByID(c, db) })
			customerRoutes.GET("/:id/reservations", func(c *gin.Context) { controllers.GetReservationsByCustomerID	(c, db) })
			customerRoutes.POST("", func(c *gin.Context) { controllers.CreateCustomer(c, db) })
			customerRoutes.PUT("/:id", func(c *gin.Context) { controllers.UpdateCustomer(c, db) })
			customerRoutes.DELETE("/:id", func(c *gin.Context) { controllers.DeleteCustomer(c, db) })
		}
		// MenuItem routes
		menuItemRoutes := api.Group("/menus")
		{
			menuItemRoutes.GET("", func(c *gin.Context) { controllers.GetMenuItems(c, db) })
			menuItemRoutes.GET("/:id", func(c *gin.Context) { controllers.GetMenuItemByID(c, db) })
			menuItemRoutes.POST("", func(c *gin.Context) { controllers.CreateMenuItem(c, db) })
			menuItemRoutes.PUT("/:id", func(c *gin.Context) { controllers.UpdateMenuItem(c, db) })
			menuItemRoutes.PATCH("/:id/availability", func(c *gin.Context) { controllers.UpdateMenuItemAvailability(c, db) })
			menuItemRoutes.DELETE("/:id", func(c *gin.Context) { controllers.DeleteMenuItem(c, db) })

		}
		// Reservation routes
		reservationRoutes := api.Group("/reservations")
		{
			reservationRoutes.GET("", func(c *gin.Context) { controllers.GetReservations(c, db) })
			reservationRoutes.GET("/:id", func(c *gin.Context) { controllers.GetReservationByID(c, db) })
			reservationRoutes.POST("", func(c *gin.Context) { controllers.CreateReservation(c, db) })
			reservationRoutes.PUT("/:id", func(c *gin.Context) { controllers.UpdateReservation(c, db) })
			reservationRoutes.PATCH("/:id/status", func(c *gin.Context) { controllers.UpdateReservationStatus(c, db) })
			reservationRoutes.DELETE("/:id", func(c *gin.Context) { controllers.DeleteReservation(c, db) })
			reservationRoutes.GET("/:id/total", func(c *gin.Context) { controllers.GetTotalByReservationID(c, db) })
		}

		// ReservationOrder routes
		orderRoutes := api.Group("/reservations/:id/orders")
		{
			orderRoutes.GET("", func(c *gin.Context) { controllers.GetOrdersByReservationID(c, db) })
			orderRoutes.POST("", func(c *gin.Context) { controllers.CreateOrder(c, db) })
			orderRoutes.PUT("/:order_id", func(c *gin.Context) { controllers.UpdateOrder(c, db) })
			orderRoutes.DELETE("/:order_id", func(c *gin.Context) { controllers.DeleteOrder(c, db) })

		}

		// Tables routes
		tableRoutes := api.Group("/tables")
		{
			tableRoutes.GET("", func(c *gin.Context) { controllers.GetTables(c, db) })
			tableRoutes.GET("/:id", func(c *gin.Context) { controllers.GetTablesByID(c, db) })
			tableRoutes.GET("/available", func(c *gin.Context) { controllers.GetAvailableTables(c, db) })
			tableRoutes.POST("", func(c *gin.Context) { controllers.CreateTable(c, db) })
			tableRoutes.PUT("/:id", func(c *gin.Context) { controllers.UpdateTable(c, db) })
			tableRoutes.PATCH("/:id/status", func(c *gin.Context) { controllers.UpdateTableStatus(c, db) })
			tableRoutes.DELETE("/:id", func(c *gin.Context) { controllers.DeleteTable(c, db) })
		}
	}

	return router
}
