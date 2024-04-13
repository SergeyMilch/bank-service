package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SergeyMilch/bank-service/internal/db"
	"github.com/SergeyMilch/bank-service/internal/handler"
	"github.com/SergeyMilch/bank-service/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func init() {
    // Загрузка переменных окружения из файла .env
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

func main() {
    // Инициализация логгера zap
    logger, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("Can't initialize zap logger: %v", err)
    }
    zap.ReplaceGlobals(logger) // Замена глобального логгера на zap
    defer logger.Sync()

    dbPool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        zap.L().Fatal("Unable to connect to database", zap.Error(err))
    }
    defer dbPool.Close()

    dbService := db.NewDB(dbPool)

    e := echo.New()

    e.Use(middleware.UserRoleMiddleware)

    bankHandler := handler.NewBankHandler(dbService, logger)
    e.POST("/bank", bankHandler.HandleBankRequest)

    // graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    server := &http.Server{
        Addr:    ":3000",
        Handler: e,
    }

    go func() {
        zap.L().Info("Starting server on port 3000")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            zap.L().Fatal("shutting down the server", zap.Error(err))
        }
    }()

    <-quit
    zap.L().Info("Shutting down server...")

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        zap.L().Fatal("Server forced to shutdown:", zap.Error(err))
    }

    zap.L().Info("Server exiting")
}
