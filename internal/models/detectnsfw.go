package models

// DetectionResult represents a single detection result from the API
type DetectionResult struct {
	Box   []int   `json:"box"`
	Class string  `json:"class"`
	Score float64 `json:"score"`
}

// APIResponse represents the response from the external NSFW detection API
type APIResponse struct {
	Filename string            `json:"filename"`
	Results  []DetectionResult `json:"results"`
	Status   string            `json:"status"`
}

// StatisticDocument represents the document structure for statistics collection
type StatisticDocument struct {
	UserID      string                    `firestore:"userId"`
	Date        string                    `firestore:"date"`
	GrandTotal  int                       `firestore:"grandTotal"`
	TotalLow    int                       `firestore:"totalLow"`
	TotalMedium int                       `firestore:"totalMedium"`
	TotalHigh   int                       `firestore:"totalHigh"`
	AppCounts   map[string]AppStatCounter `firestore:"appCounts"`
}

// AppStatCounter represents counter for each application
type AppStatCounter struct {
	Total  int `firestore:"total"`
	Low    int `firestore:"low"`
	Medium int `firestore:"medium"`
	High   int `firestore:"high"`
}
