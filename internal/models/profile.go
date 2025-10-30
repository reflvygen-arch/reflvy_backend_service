// internal/models/models.go
package models

// ... ProfileResponse yang sudah ada
type ProfileResponse struct {
	Message    string `json:"message"`
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	DisplayName string `json:"display_name,omitempty"` // Tambahkan displayName
	IsVerified bool   `json:"is_verified"`
	Gender     string `json:"gender,omitempty"` // Tambahkan gender dan age
	Age        int    `json:"age,omitempty"`
}

// Model untuk menyimpan data tambahan
type UserDetails struct {
	Gender string `json:"gender" binding:"required"`
	Age    int    `json:"age" binding:"required"`
	Email  string `json:"email"` // Simpan juga email untuk kemudahan query
}
