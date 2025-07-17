package main

import (
	"log"
	"os"

	"github.com/azainwork/core-banking-api/db"
	"github.com/azainwork/core-banking-api/middleware"
	"github.com/azainwork/core-banking-api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	database, err := db.InitDB()
	if err != nil {
		logger.Fatal("Failed to connect to database:", err)
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger(logger))

	routes.SetupRoutes(router, database)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	logger.Info("Starting Core Banking API server on port " + port)
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server:", err)
	}
}
