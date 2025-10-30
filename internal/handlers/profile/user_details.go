// internal/handlers/profile/user_details.go
package profile

import (
	"context"
	"log"
	"net/http"

	"go-gin-project/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func SaveUserDetailsHandler(db *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.MustGet("uid").(string)
		email := c.MustGet("email").(string)

		var userDetails models.UserDetails
		if err := c.ShouldBindJSON(&userDetails); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Set email dari token agar aman
		userDetails.Email = email

		// Simpan data ke Firestore dengan UID sebagai ID dokumen
		// Ini akan membuat collection 'users' jika belum ada
		_, err := db.Collection("users").Doc(uid).Set(context.Background(), userDetails)
		if err != nil {
			// TAMBAHKAN BARIS INI untuk melihat error asli di terminal
			log.Printf("Error saving to Firestore: %v\n", err)

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user details"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User details saved successfully"})
	}
}
