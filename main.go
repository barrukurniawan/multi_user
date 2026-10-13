package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/barrukurniawan/multi_user/core/config"
	"github.com/barrukurniawan/multi_user/core/controllers"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/barrukurniawan/multi_user/core/repositories"
	"github.com/barrukurniawan/multi_user/core/routers"
	"github.com/barrukurniawan/multi_user/core/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		fatal("failed to load config", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL()), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fatal("failed to connect to database", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fatal("failed to get database connection", err)
	}
	defer sqlDB.Close()

	slog.Info("database connected", "db", cfg.DBName)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fatal("failed to connect to redis", err)
	}
	defer rdb.Close()

	slog.Info("Redis connected", "redis", cfg.RedisAddr)

	if err := runMigrations(db, "migrations"); err != nil {
		fatal("failed to run migrations", err)
	}

	userRepo := repositories.NewUserRepository(db)
	jwtManager := helpers.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryHours)
	authService := services.NewAuthService(userRepo, jwtManager)
	authController := controllers.NewAuthController(authService)

	taskRepo := repositories.NewTaskRepository(db)
	taskLogRepo := repositories.NewTaskLogRepository()
	idemRepo := repositories.NewIdempotencyRepository(rdb)
	dbManager := repositories.NewDbManager(db)
	notifier := services.NewNotificationService()
	taskService := services.NewTaskService(taskRepo, taskLogRepo, idemRepo, userRepo, dbManager, notifier)
	taskController := controllers.NewTaskController(taskService)

	r := gin.New()
	r.Use(helpers.LoggingMiddleware(), helpers.RecoveryMiddleware())

	routerConfig := routers.Config{
		Server:         r,
		JWTManager:     jwtManager,
		AuthController: authController,
		TaskController: taskController,
	}
	routerConfig.Setup()

	slog.Info("server started", "port", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		fatal("server stopped", err)
	}
}

// Log ERROR (slog tidak punya level Fatal)
func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}

func runMigrations(db *gorm.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if err := db.Exec(string(sql)).Error; err != nil {
			return err
		}
		slog.Info("migration applied", "file", file)
	}
	return nil
}
