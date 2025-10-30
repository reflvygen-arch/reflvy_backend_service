// internal/handlers/statistic/dummy.go
package statistic

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"go-gin-project/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

var dummyApps = []string{"tiktok", "chrome", "gallery", "instagram", "youtube", "facebook", "twitter"}

// GenerateDummyStatisticHandler generates dummy statistics for a specific email (historical data)
func GenerateDummyStatisticHandler(db *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := "dummyuser@gmail.com" // bisa diubah ke multi user jika mau
		startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Now()
		numDays := int(endDate.Sub(startDate).Hours()/24) + 1

		for i := 0; i < numDays; i++ {
			date := startDate.AddDate(0, 0, i)
			dateString := date.Format("January 2, 2006")
			emailPart := userId[:len(userId)-10] // ambil sebelum @, asumsi @gmail.com
			docID := emailPart + "_" + date.Format("2006-01-02")

			appCounts := make(map[string]models.AppStatCounter)
			grandTotal := 0
			totalLow := 0
			totalMedium := 0
			totalHigh := 0

			// random jumlah aplikasi per hari (3-6)
			appCount := rand.Intn(4) + 3
			usedApps := rand.Perm(len(dummyApps))[:appCount]
			for _, idx := range usedApps {
				app := dummyApps[idx]
				low := rand.Intn(10)
				medium := rand.Intn(5)
				high := rand.Intn(3)
				total := low + medium + high
				appCounts[app] = models.AppStatCounter{
					Total:  total,
					Low:    low,
					Medium: medium,
					High:   high,
				}
				grandTotal += total
				totalLow += low
				totalMedium += medium
				totalHigh += high
			}

			statDoc := models.StatisticDocument{
				UserID:      userId,
				Date:        dateString,
				GrandTotal:  grandTotal,
				TotalLow:    totalLow,
				TotalMedium: totalMedium,
				TotalHigh:   totalHigh,
				AppCounts:   appCounts,
			}

			_, err := db.Collection("nsfw_stats").Doc(docID).Set(context.Background(), statDoc)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed at " + docID, "detail": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Dummy statistics generated successfully"})
	}
}

// GenerateTodayDummyStatisticHandler generates dummy statistics for specified period with email input (POST only)
func GenerateTodayDummyStatisticHandler(db *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only accept POST method with JSON body
		var reqBody struct {
			Email  string `json:"email" binding:"required,email"`
			Period string `json:"period"`
		}

		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. Expected JSON with email and optional period"})
			return
		}

		email := reqBody.Email
		period := reqBody.Period

		// Default period is today if not specified
		if period == "" {
			period = "today"
		}

		// Validate period
		validPeriods := []string{"today", "7days", "1month", "3months"}
		isValid := false
		for _, validPeriod := range validPeriods {
			if period == validPeriod {
				isValid = true
				break
			}
		}
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Options: today, 7days, 1month, 3months"})
			return
		}

		// Validasi email sederhana
		if len(email) < 5 || !strings.Contains(email, "@") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		// Calculate date range based on period
		now := time.Now()
		var startDate time.Time

		switch period {
		case "today":
			startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		case "7days":
			startDate = now.AddDate(0, 0, -6) // 7 days including today
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		case "1month":
			startDate = now.AddDate(0, -1, 0) // 1 month ago
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		case "3months":
			startDate = now.AddDate(0, -3, 0) // 3 months ago
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		}

		emailPart := email[:strings.Index(email, "@")] // ambil sebelum @

		// Track statistics
		createdCount := 0
		skippedCount := 0
		var createdDates []string
		var skippedDates []string

		// Loop through dates in range
		current := startDate
		for current.Before(now.AddDate(0, 0, 1)) || current.Equal(now) {
			dateString := current.Format("January 2, 2006")
			docID := emailPart + "_" + current.Format("2006-01-02")

			// Check if document already exists
			doc, err := db.Collection("nsfw_stats").Doc(docID).Get(context.Background())
			if err == nil && doc.Exists() {
				// Document already exists, skip
				skippedCount++
				skippedDates = append(skippedDates, dateString)
				current = current.AddDate(0, 0, 1)
				continue
			}

			// Create new dummy data for this date
			appCounts := make(map[string]models.AppStatCounter)
			grandTotal := 0
			totalLow := 0
			totalMedium := 0
			totalHigh := 0

			// random jumlah aplikasi per hari (2-5)
			appCount := rand.Intn(4) + 2
			usedApps := rand.Perm(len(dummyApps))[:appCount]
			for _, idx := range usedApps {
				app := dummyApps[idx]
				low := rand.Intn(8) + 1 // 1-8 low detections
				medium := rand.Intn(4)  // 0-3 medium detections
				high := rand.Intn(2)    // 0-1 high detections
				total := low + medium + high
				appCounts[app] = models.AppStatCounter{
					Total:  total,
					Low:    low,
					Medium: medium,
					High:   high,
				}
				grandTotal += total
				totalLow += low
				totalMedium += medium
				totalHigh += high
			}

			statDoc := models.StatisticDocument{
				UserID:      email,
				Date:        dateString,
				GrandTotal:  grandTotal,
				TotalLow:    totalLow,
				TotalMedium: totalMedium,
				TotalHigh:   totalHigh,
				AppCounts:   appCounts,
			}

			_, err = db.Collection("nsfw_stats").Doc(docID).Set(context.Background(), statDoc)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  "Failed to create dummy data at " + docID,
					"detail": err.Error(),
				})
				return
			}

			createdCount++
			createdDates = append(createdDates, dateString)

			current = current.AddDate(0, 0, 1)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Dummy statistics generation completed",
			"email":        email,
			"period":       period,
			"startDate":    startDate.Format("January 2, 2006"),
			"endDate":      now.Format("January 2, 2006"),
			"createdCount": createdCount,
			"skippedCount": skippedCount,
			"createdDates": createdDates,
			"skippedDates": skippedDates,
		})
	}
}
