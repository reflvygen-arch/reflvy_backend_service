package detectnsfw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"go-gin-project/internal/models"
	"go-gin-project/internal/services"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func DetectNSFWHandler(db *firestore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse multipart form
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
			return
		}
		defer file.Close()

		// Get application parameter from form
		application := c.PostForm("application")
		if application == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Application parameter is required"})
			return
		}

		// Get user email from context (set by auth middleware)
		email, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User email not found in context"})
			return
		}
		userEmail := email.(string)

		// Read the file content
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image file"})
			return
		}

		// Create a new multipart form for forwarding
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add the image file to the new form
		part, err := writer.CreateFormFile("image", header.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form file"})
			return
		}
		part.Write(fileBytes)
		writer.Close()

		// Forward the request to the external service
		req, err := http.NewRequest("POST", "http://127.0.0.1:5000/detect", body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward request"})
			return
		}
		defer resp.Body.Close()

		// Read the response from the external service
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Parse the JSON response
		var apiResp models.APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API response"})
			return
		}

		// Classify NSFW level
		nsfwLevel := services.ClassifyNSFW(apiResp.Results)

		// If NSFW level > 0, save to Firestore with proper document naming and counting
		if nsfwLevel > 0 {
			err = updateStatisticDocument(db, userEmail, application, nsfwLevel)
			if err != nil {
				// Log error but don't fail the request
				// You could add proper logging here
			}
		}

		// Return the classification result along with original detection results
		c.JSON(http.StatusOK, gin.H{
			"filename":          apiResp.Filename,
			"nsfw_level":        nsfwLevel,
			"detection_results": apiResp.Results,
			"status":            "success",
		})
	}
}

// updateStatisticDocument creates or updates the statistic document with proper counting and app tracking
func updateStatisticDocument(db *firestore.Client, email, application string, nsfwLevel int) error {
	now := time.Now()

	// Create document ID in format: email_YYYY-MM-DD
	emailPart := strings.Split(email, "@")[0] // Get part before @
	docID := fmt.Sprintf("%s_%04d-%02d-%02d", emailPart, now.Year(), int(now.Month()), now.Day())

	// Format date as "September 19, 2025" (no time)
	dateString := now.Format("January 2, 2006")

	docRef := db.Collection("nsfw_stats").Doc(docID)

	// Try to get existing document
	doc, err := docRef.Get(context.Background())

	if err != nil {
		// Document doesn't exist, create new one
		appCounts := make(map[string]models.AppStatCounter)

		// Initialize app counter
		appCounter := models.AppStatCounter{
			Total:  1,
			Low:    0,
			Medium: 0,
			High:   0,
		}

		// Set appropriate counter based on NSFW level
		switch nsfwLevel {
		case 1:
			appCounter.Low = 1
		case 2:
			appCounter.Medium = 1
		case 3:
			appCounter.High = 1
		}

		appCounts[strings.ToLower(application)] = appCounter

		newDoc := models.StatisticDocument{
			UserID:      email,
			Date:        dateString,
			GrandTotal:  1,
			TotalLow:    0,
			TotalMedium: 0,
			TotalHigh:   0,
			AppCounts:   appCounts,
		}

		// Set grand totals
		switch nsfwLevel {
		case 1:
			newDoc.TotalLow = 1
		case 2:
			newDoc.TotalMedium = 1
		case 3:
			newDoc.TotalHigh = 1
		}

		_, err = docRef.Set(context.Background(), newDoc)
		return err
	} else {
		// Document exists, update counters
		var existingDoc models.StatisticDocument
		err = doc.DataTo(&existingDoc)
		if err != nil {
			return err
		}

		// Initialize appCounts if nil
		if existingDoc.AppCounts == nil {
			existingDoc.AppCounts = make(map[string]models.AppStatCounter)
		}

		appKey := strings.ToLower(application)

		// Get existing app counter or create new one
		appCounter, exists := existingDoc.AppCounts[appKey]
		if !exists {
			appCounter = models.AppStatCounter{
				Total:  0,
				Low:    0,
				Medium: 0,
				High:   0,
			}
		}

		// Update app-specific counters
		appCounter.Total++
		switch nsfwLevel {
		case 1:
			appCounter.Low++
			existingDoc.TotalLow++
		case 2:
			appCounter.Medium++
			existingDoc.TotalMedium++
		case 3:
			appCounter.High++
			existingDoc.TotalHigh++
		}

		// Update grand total
		existingDoc.GrandTotal++

		// Save updated app counter
		existingDoc.AppCounts[appKey] = appCounter

		_, err = docRef.Set(context.Background(), existingDoc)
		return err
	}
}
