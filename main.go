package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/barrukurniawan/multi_user/core/config"
	"github.com/barrukurniawan/multi_user/core/routers"
	"github.com/gin-gonic/gin"
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
		TranslateError: true,                                  // error UNIQUE jadi gorm.ErrDuplicatedKey
		Logger:         logger.Default.LogMode(logger.Silent), // log SQL bawaan GORM bukan JSON, jadi dimatikan
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

	if err := runMigrations(db, "migrations"); err != nil {
		fatal("failed to run migrations", err)
	}

	r := gin.Default()
	routerConfig := routers.Config{
		Server: r,
	}
	routerConfig.Setup()
	r.Run(":8080")
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
