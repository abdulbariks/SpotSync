package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"spotsync/handler"
	"spotsync/middleware"
	"spotsync/models"
	"spotsync/repository"
	"spotsync/service"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	var dsn string
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		dsn = databaseURL
	} else {
		dsn = "host=" + getEnv("DB_HOST", "localhost") +
			" port=" + getEnv("DB_PORT", "5432") +
			" user=" + getEnv("DB_USER", "postgres") +
			" password=" + getEnv("DB_PASSWORD", "postgres") +
			" dbname=" + getEnv("DB_NAME", "spotsync") +
			" sslmode=" + getEnv("DB_SSLMODE", "disable") +
			" TimeZone=UTC"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database connection:", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := db.AutoMigrate(&models.User{}, &models.ParkingZone{}, &models.Reservation{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "default-secret-key")
	jwtExpiryHours := 24
	if h := os.Getenv("JWT_EXPIRY_HOURS"); h != "" {
		jwtExpiryHours, _ = strconv.Atoi(h)
	}

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, jwtSecret, jwtExpiryHours)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	h := handler.NewHandler(svc)
	jwtMiddleware := middleware.NewJWTMiddleware(svc)

	e.POST("/api/v1/auth/register", h.Register)
	e.POST("/api/v1/auth/login", h.Login)

	e.GET("/api/v1/zones", h.GetAllZones)
	e.GET("/api/v1/zones/:id", h.GetZone)

	adminGroup := e.Group("/api/v1")
	adminGroup.Use(jwtMiddleware.VerifyJWT, adminOnly)
	adminGroup.POST("/zones", h.CreateZone)
	adminGroup.PUT("/zones/:id", h.UpdateZone)
	adminGroup.DELETE("/zones/:id", h.DeleteZone)
	adminGroup.GET("/reservations", h.GetAllReservations)

	authGroup := e.Group("/api/v1")
	authGroup.Use(jwtMiddleware.VerifyJWT)
	authGroup.POST("/reservations", h.CreateReservation)
	authGroup.GET("/reservations/my-reservations", h.GetMyReservations)
	authGroup.DELETE("/reservations/:id", h.CancelReservation)

	port := getEnv("SERVER_PORT", "8080")
	log.Printf("Server starting on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func adminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := c.Get("user").(*models.JWTClaims)
		if !ok || user.Role != models.RoleAdmin {
			return c.JSON(http.StatusForbidden, echo.Map{"success": false, "message": "Admin access required"})
		}
		return next(c)
	}
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}