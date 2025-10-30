// internal/handlers/profile/profile.go

package profile

import (
	"context"
	"net/http"

	"go-gin-project/internal/models"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func ProfileHandler(authClient *auth.Client, db *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.MustGet("uid").(string)
		email := c.MustGet("email").(string)
		isVerified := c.MustGet("is_verified").(bool)

		// Ambil displayName dari Firebase Auth
		var displayName string
		user, err := authClient.GetUser(context.Background(), uid)
		if err == nil && user != nil {
			displayName = user.DisplayName
		}

		// Ambil data tambahan dari Firestore
		doc, err := db.Collection("users").Doc(uid).Get(context.Background())

		var userDetails models.UserDetails
		if err == nil {
			// Jika dokumen ditemukan, map data ke struct
			doc.DataTo(&userDetails)
		}

		response := models.ProfileResponse{
			Message:     "Welcome " + email + "!",
			UserID:      uid,
			Email:       email,
			DisplayName: displayName, // Tambahkan displayName dari Firebase Auth
			IsVerified:  isVerified,
			Gender:      userDetails.Gender, // Tambahkan data dari Firestore
			Age:         userDetails.Age,    // Tambahkan data dari Firestore
		}

		c.JSON(http.StatusOK, response)
	}
}
