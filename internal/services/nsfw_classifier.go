package services

import "go-gin-project/internal/models"

// ClassifyNSFW classifies the NSFW level based on detection results using updated standards
func ClassifyNSFW(results []models.DetectionResult) int {
	// Initialize counters for exposed areas
	exposedCount := 0
	hasHighExposed := false

	femaleBreastExposed := 0.0
	femaleGenitaliaExposed := 0.0
	maleGenitaliaExposed := 0.0
	anusExposed := 0.0
	bellyExposed := 0.0
	buttocksExposed := 0.0
	armpitsExposed := 0.0
	feetExposed := 0.0
	femaleBreastCovered := 1.0 // Default high if not detected

	for _, r := range results {
		switch r.Class {
		case "FEMALE_BREAST_EXPOSED":
			femaleBreastExposed = r.Score
			if r.Score >= 0.5 {
				hasHighExposed = true
			}
			if r.Score >= 0.2 {
				exposedCount++
			}
		case "FEMALE_GENITALIA_EXPOSED":
			femaleGenitaliaExposed = r.Score
			if r.Score >= 0.5 {
				hasHighExposed = true
			}
			if r.Score >= 0.2 {
				exposedCount++
			}
		case "MALE_GENITALIA_EXPOSED":
			maleGenitaliaExposed = r.Score
			if r.Score >= 0.5 {
				hasHighExposed = true
			}
			if r.Score >= 0.2 {
				exposedCount++
			}
		case "ANUS_EXPOSED":
			anusExposed = r.Score
			if r.Score >= 0.5 {
				hasHighExposed = true
			}
			if r.Score >= 0.2 {
				exposedCount++
			}
		case "BELLY_EXPOSED":
			bellyExposed = r.Score
			if r.Score >= 0.3 {
				exposedCount++
			}
		case "BUTTOCKS_EXPOSED":
			buttocksExposed = r.Score
			if r.Score >= 0.3 {
				exposedCount++
			}
		case "ARMPITS_EXPOSED":
			armpitsExposed = r.Score
			if r.Score >= 0.3 {
				exposedCount++
			}
		case "FEET_EXPOSED":
			feetExposed = r.Score
		case "FEMALE_BREAST_COVERED":
			if r.Score < femaleBreastCovered {
				femaleBreastCovered = r.Score
			}
		}
	}

	// Check for category 4: high NSFW (explicit)
	if hasHighExposed || exposedCount > 2 {
		return 3
	}

	// Check for category 3: moderate NSFW (minimal clothing)
	if buttocksExposed >= 0.5 || (bellyExposed >= 0.5 && femaleBreastCovered < 0.4) {
		if femaleBreastExposed < 0.5 && femaleGenitaliaExposed < 0.3 && maleGenitaliaExposed < 0.3 && anusExposed < 0.3 {
			return 2
		}
	}

	// Check for category 2: mild NSFW (casual sensual)
	if (bellyExposed >= 0.5 || armpitsExposed >= 0.5 || feetExposed >= 0.5) && femaleBreastCovered >= 0.4 {
		if femaleBreastExposed < 0.3 && femaleGenitaliaExposed < 0.3 && maleGenitaliaExposed < 0.3 && anusExposed < 0.3 {
			return 1
		}
	}

	// Check for category 1: not NSFW (safe)
	if femaleBreastExposed < 0.2 && femaleGenitaliaExposed < 0.2 && maleGenitaliaExposed < 0.2 && anusExposed < 0.2 &&
		bellyExposed < 0.3 && buttocksExposed < 0.3 && armpitsExposed < 0.3 {
		return 0
	}

	// Default to mild if no clear category
	return 1
}
