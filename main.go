// main.go

package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	"go-gin-project/internal/routes"
)

var (
	router          *gin.Engine
	authClient      *auth.Client
	firestoreClient *firestore.Client
)

func init() {
	// Load environment variables from .env file (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize Firebase
	authClient, firestoreClient = setupFirebase()

	// Initialize Gin router
	router = gin.Default()
	routes.SetupRoutes(router, authClient, firestoreClient)
}

// setupFirebase initializes Firebase Admin SDK and returns auth & firestore clients
func setupFirebase() (*auth.Client, *firestore.Client) {
	var opt option.ClientOption

	// Try to load from JSON string first (for Vercel deployment)
	credJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if credJSON != "" {
		opt = option.WithCredentialsJSON([]byte(credJSON))
	} else {
		// Fallback to file path (for local development)
		credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
		if credPath == "" {
			log.Fatal("Neither FIREBASE_CREDENTIALS_JSON nor FIREBASE_CREDENTIALS_PATH environment variable is set")
		}
		opt = option.WithCredentialsFile(credPath)
	}

	// Inisialisasi App
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}

	// Inisialisasi Auth Client
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error getting Firebase Auth client: %v\n", err)
	}

	// Inisialisasi Firestore Client
	firestoreClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firestore client: %v", err)
	}

	return authClient, firestoreClient
}

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// Handler is the exported serverless function handler for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

func main() {
	// For local development
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	address := host + ":" + port

	log.Printf("Server running on http://%s", address)
	if host == "0.0.0.0" {
		log.Printf("Local network access: http://%s:%s", getLocalIP(), port)
		log.Printf("Localhost access: http://localhost:%s", port)
	} else if host == "localhost" || host == "127.0.0.1" {
		log.Printf("Localhost access only: http://localhost:%s", port)
	}

	router.Run(address)
}
