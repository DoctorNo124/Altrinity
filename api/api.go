package main

import (
	"log"
	"os"
	"time"

	"altrinity/api/controllers"
	"altrinity/api/repositories"
	"altrinity/api/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file into environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
}

func main() {
	keycloakURL = os.Getenv("KEYCLOAK_URL")    // e.g. http://localhost:8080
	realm = os.Getenv("KEYCLOAK_REALM")        // e.g. my-app
	clientID = os.Getenv("KEYCLOAK_CLIENT_ID") // e.g. go-api
	clientSecret = os.Getenv("KEYCLOAK_CLIENT_SECRET")

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://192.168.50.175:3000/", "http://192.168.56.1:3000/", "http://172.27.240.1:3000/", "https://altrinity.com", "https://app.altrinitytech.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- Setup dependencies ---
	repo := &repositories.KeycloakRepo{
		BaseURL:      keycloakURL,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	service := &services.AdminService{
		Repo: repo,
	}

	adminController := &controllers.AdminController{
		Service: service,
	}

	api := r.Group("/api")
	{
		api.GET("/users", AuthMiddleware("admin"), adminController.ListUsers)
		api.PUT("/users/:id/role", AuthMiddleware("admin"), adminController.UpdateUserRole)
		api.GET("/public", func(c *gin.Context) {
			c.JSON(200, gin.H{"msg": "anyone can access this"})
		})
		api.GET("/user", AuthMiddleware(""), func(c *gin.Context) {
			c.JSON(200, gin.H{"msg": "hello user"})
		})
		api.GET("/admin", AuthMiddleware("admin"), func(c *gin.Context) {
			c.JSON(200, gin.H{"msg": "hello admin"})
		})
	}

	r.Run("0.0.0.0:8081")
}
