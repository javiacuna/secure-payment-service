package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"secure-payment-service/internal/config"
	"secure-payment-service/internal/controller"
	"secure-payment-service/internal/logging"
	"secure-payment-service/internal/middleware"
	"secure-payment-service/internal/models"
	"secure-payment-service/internal/repository"
	"secure-payment-service/internal/routes"
	"secure-payment-service/internal/service"
)

func main() {
	logging.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		logging.Logger.Fatalf("Failed to load configuration: %v", err)
	}
	logging.Logger.WithFields(logrus.Fields{
		"database_url": cfg.DatabaseURL,
		"address":      cfg.Address,
	}).Info("Configuration loaded successfully")

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		logging.Logger.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&models.Transfer{})
	if err != nil {
		logging.Logger.Fatalf("Failed to auto migrate database: %v", err)
	}
	logging.Logger.Info("Database connection established and migrations run successfully.")

	repo := repository.NewGormRepository(db)
	svc := service.NewTransferService(repo)
	ctrl := controller.NewTransferController(svc)

	router := gin.Default()

	jwtMiddleware := middleware.Auth()

	routes.SetupRoutes(router, jwtMiddleware, ctrl)

	logging.Logger.WithField("address", cfg.Address).Info("Server running")
	logging.Logger.Fatal(router.Run(cfg.Address))
}
