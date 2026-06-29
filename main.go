package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"spotsync/reservation"
	"spotsync/zone"
	"spotsync/user"

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

	if err := db.AutoMigrate(&user.User{}, &zone.ParkingZone{}, &reservation.Reservation{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "default-secret-key")
	jwtExpiryHours := 24
	if h := os.Getenv("JWT_EXPIRY_HOURS"); h != "" {
		jwtExpiryHours, _ = strconv.Atoi(h)
	}

	userRepo := user.NewRepository(db)
	zoneRepo := zone.NewRepository(db)
	reservationRepo := reservation.NewRepository(db)

	userSvc := user.NewService(userRepo, jwtSecret, jwtExpiryHours)
	zoneSvc := zone.NewService(zoneRepo)
	reservationSvc := reservation.NewService(reservationRepo)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	userHandler := user.NewHTTPHandler(userSvc)
	zoneHandler := zone.NewHTTPHandler(zoneSvc)
	reservationHandler := reservation.NewHTTPHandler(reservationSvc)
	jwtMiddleware := user.NewJWTMiddleware(userSvc)

	e.POST("/api/v1/auth/register", userHandler.Register)
	e.POST("/api/v1/auth/login", userHandler.Login)

	e.GET("/api/v1/zones", zoneHandler.GetAllZones)
	e.GET("/api/v1/zones/:id", zoneHandler.GetZone)

	adminGroup := e.Group("/api/v1")
	adminGroup.Use(jwtMiddleware.VerifyJWT, user.AdminOnly)
	adminGroup.POST("/zones", zoneHandler.CreateZone)
	adminGroup.PUT("/zones/:id", zoneHandler.UpdateZone)
	adminGroup.DELETE("/zones/:id", zoneHandler.DeleteZone)
	adminGroup.GET("/reservations", reservationHandler.GetAllReservations)

	authGroup := e.Group("/api/v1")
	authGroup.Use(jwtMiddleware.VerifyJWT)
	authGroup.POST("/reservations", reservationHandler.CreateReservation)
	authGroup.GET("/reservations/my-reservations", reservationHandler.GetMyReservations)
	authGroup.DELETE("/reservations/:id", reservationHandler.CancelReservation)

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

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}