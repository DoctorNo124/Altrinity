package main

import (
	"altrinity/api/controllers"
	"altrinity/api/middleware"
	"altrinity/api/repositories"
	"altrinity/api/services"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	// Load .env file into environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
}

func main() {
	middleware.InitJWKS()
	keycloakURL := os.Getenv("KEYCLOAK_URL")    // e.g. http://localhost:8080
	realm := os.Getenv("KEYCLOAK_REALM")        // e.g. my-app
	clientID := os.Getenv("KEYCLOAK_CLIENT_ID") // e.g. go-api
	clientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")
	db, err := sqlx.Connect("postgres", os.Getenv("POSTGRES_DSN"))
	if err != nil {
		log.Fatal("DB connect error:", err)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://app.altrinitytech.com", "http://localhost:3000"},
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

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"), // e.g. localhost:6379
	})

	volRepo := &repositories.VolunteerRepository{DB: db, Redis: redisClient}
	volService := &services.VolunteerService{Repo: volRepo}
	volController := &controllers.VolunteerController{Service: volService}

	api := r.Group("/api")
	{
		api.GET("/users", middleware.AuthMiddleware("admin"), adminController.ListUsers)
		api.PUT("/users/:id/role", middleware.AuthMiddleware("admin"), adminController.UpdateUserRole)
		api.GET("/public", func(c *gin.Context) {
			c.JSON(200, gin.H{"msg": "anyone can access this"})
		})
		api.GET("/user", middleware.AuthMiddleware(""), func(c *gin.Context) {
			c.JSON(200, gin.H{"msg": "hello user"})
		})
		api.GET("/admin", middleware.AuthMiddleware("admin"), func(c *gin.Context) {
			c.JSON(200, gin.H{"msg": "hello admin"})
		})
		api.POST("/positions", middleware.AuthMiddleware("volunteer"), volController.UpdatePosition)
		api.GET("/positions", middleware.AuthMiddleware("admin"), volController.GetPositions)
		api.GET("/ws/positions", volController.StreamPositions)
	}

	r.Run("0.0.0.0:8081")
}
