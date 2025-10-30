// internal/routes/routes.go

package routes

import (
	"github.com/gin-gonic/gin"

	"go-gin-project/internal/handlers/detectnsfw"
	"go-gin-project/internal/handlers/profile"
	"go-gin-project/internal/handlers/statistic"
	"go-gin-project/internal/middleware"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(router *gin.Engine, authClient *auth.Client, db *firestore.Client) {
	// Public routes
	router.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "This is a public endpoint"})
	})

	// Route untuk generate dummy statistik hari ini dengan email input (tidak perlu auth, hanya untuk dev)
	router.POST("/api/statistic/dummy", statistic.GenerateTodayDummyStatisticHandler(db))

	// Route untuk generate dummy statistik historis (tidak perlu auth, hanya untuk dev)
	router.POST("/api/statistic/dummy/historical", statistic.GenerateDummyStatisticHandler(db))

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(authClient))
	{
		// Endpoint lama untuk mendapatkan profil (akan kita update)
		protected.GET("/profile", profile.ProfileHandler(authClient, db)) // Berikan authClient dan db client ke handler

		// Endpoint BARU untuk menyimpan detail gender dan usia
		protected.POST("/profile/details", profile.SaveUserDetailsHandler(db))

		// Endpoint untuk detect NSFW
		protected.POST("/detectnsfw", detectnsfw.DetectNSFWHandler(db))

		// Endpoint untuk mendapatkan statistik berdasarkan periode
		protected.GET("/statistics", statistic.GetStatisticHandler(db))
	}
}
