package statistic

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go-gin-project/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

// GetStatisticHandler handles requests for statistic data based on time period
func GetStatisticHandler(db *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get period parameter from query
		period := c.Query("period")
		if period == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Period parameter is required. Options: today, 7days, 1month, 3months"})
			return
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

		// Get user email from context (set by auth middleware)
		email, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User email not found in context"})
			return
		}
		userEmail := email.(string)

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

		// Get email part for document ID prefix
		emailPart := strings.Split(userEmail, "@")[0]

		// Query Firestore for user's statistics within date range
		stats, err := getStatisticsInDateRange(db, emailPart, startDate, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics", "detail": err.Error()})
			return
		}

		// Aggregate statistics
		aggregatedStats := aggregateStatistics(stats, period)

		c.JSON(http.StatusOK, gin.H{
			"period":     period,
			"email":      userEmail,
			"startDate":  startDate.Format("January 2, 2006"),
			"endDate":    now.Format("January 2, 2006"),
			"statistics": aggregatedStats,
			"status":     "success",
		})
	}
}

// getStatisticsInDateRange retrieves statistics documents within the specified date range
func getStatisticsInDateRange(db *firestore.Client, emailPart string, startDate, endDate time.Time) ([]models.StatisticDocument, error) {
	var stats []models.StatisticDocument

	// Generate all possible document IDs in the date range
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		docID := emailPart + "_" + current.Format("2006-01-02")

		doc, err := db.Collection("nsfw_stats").Doc(docID).Get(context.Background())
		if err != nil {
			// Document doesn't exist for this date, skip
			current = current.AddDate(0, 0, 1)
			continue
		}

		var stat models.StatisticDocument
		if err := doc.DataTo(&stat); err != nil {
			// Error parsing document, skip
			current = current.AddDate(0, 0, 1)
			continue
		}

		stats = append(stats, stat)
		current = current.AddDate(0, 0, 1)
	}

	return stats, nil
}

// DailySummary represents simplified daily statistics without app breakdown (for multi-day periods)
type DailySummary struct {
	Date        string `json:"date"`
	GrandTotal  int    `json:"grandTotal"`
	TotalLow    int    `json:"totalLow"`
	TotalMedium int    `json:"totalMedium"`
	TotalHigh   int    `json:"totalHigh"`
}

// PeriodStatistics represents comprehensive statistics for non-today periods
type PeriodStatistics struct {
	// Overall totals for the entire period
	TotalGrandTotal int `json:"totalGrandTotal"`
	TotalLow        int `json:"totalLow"`
	TotalMedium     int `json:"totalMedium"`
	TotalHigh       int `json:"totalHigh"`

	// Per-app totals for the entire period
	AppBreakdown map[string]models.AppStatCounter `json:"appBreakdown"`

	// Daily breakdown with per-app details for each day
	DailyBreakdown []DailySummary `json:"dailyBreakdown"`
}

// aggregateStatistics combines multiple daily statistics based on the period
func aggregateStatistics(stats []models.StatisticDocument, period string) interface{} {
	if len(stats) == 0 {
		if period == "today" {
			return map[string]interface{}{
				"totalGrandTotal": 0,
				"totalLow":        0,
				"totalMedium":     0,
				"totalHigh":       0,
				"appBreakdown":    map[string]models.AppStatCounter{},
			}
		} else {
			return PeriodStatistics{
				TotalGrandTotal: 0,
				TotalLow:        0,
				TotalMedium:     0,
				TotalHigh:       0,
				AppBreakdown:    map[string]models.AppStatCounter{},
				DailyBreakdown:  []DailySummary{},
			}
		}
	}

	totalGrandTotal := 0
	totalLow := 0
	totalMedium := 0
	totalHigh := 0
	appBreakdown := make(map[string]models.AppStatCounter)

	for _, stat := range stats {
		totalGrandTotal += stat.GrandTotal
		totalLow += stat.TotalLow
		totalMedium += stat.TotalMedium
		totalHigh += stat.TotalHigh

		// Aggregate app statistics
		for appName, appCounter := range stat.AppCounts {
			if existing, exists := appBreakdown[appName]; exists {
				existing.Total += appCounter.Total
				existing.Low += appCounter.Low
				existing.Medium += appCounter.Medium
				existing.High += appCounter.High
				appBreakdown[appName] = existing
			} else {
				appBreakdown[appName] = appCounter
			}
		}
	}

	if period == "today" {
		// For "today", return only totals and app breakdown (no daily breakdown)
		return map[string]interface{}{
			"totalGrandTotal": totalGrandTotal,
			"totalLow":        totalLow,
			"totalMedium":     totalMedium,
			"totalHigh":       totalHigh,
			"appBreakdown":    appBreakdown,
		}
	}

	// For other periods, return comprehensive structure without app details in daily breakdown
	var dailySummaries []DailySummary
	for _, stat := range stats {
		summary := DailySummary{
			Date:        stat.Date,
			GrandTotal:  stat.GrandTotal,
			TotalLow:    stat.TotalLow,
			TotalMedium: stat.TotalMedium,
			TotalHigh:   stat.TotalHigh,
		}
		dailySummaries = append(dailySummaries, summary)
	}

	return PeriodStatistics{
		TotalGrandTotal: totalGrandTotal,
		TotalLow:        totalLow,
		TotalMedium:     totalMedium,
		TotalHigh:       totalHigh,
		AppBreakdown:    appBreakdown,
		DailyBreakdown:  dailySummaries,
	}
}