package main

import (
	"context"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/vlks-dev/EffectiveMobileGoTest/docs"

	"github.com/vlks-dev/EffectiveMobileGoTest/database"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/handlers"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/repositories"
	"github.com/vlks-dev/EffectiveMobileGoTest/internal/services"
	"github.com/vlks-dev/EffectiveMobileGoTest/migrations"
	"github.com/vlks-dev/EffectiveMobileGoTest/utils/config"
	"github.com/vlks-dev/EffectiveMobileGoTest/utils/logger"
)

// @BasePath /music_library/v1
func main() {
	cfg := config.LoadConfig()
	slog := logger.NewSlog()
	slog.Debug("config and logger initialized",
		"configuration", cfg)

	ctx := context.Background()
	err := migrations.RunMigrations(ctx, slog, cfg)
	if err != nil {
		slog.Error("migration failed", "error", err.Error())
		return
	}
	pool, err := database.NewPostgresPool(cfg, slog, ctx)
	if err != nil {
		slog.Error("postgres pool failed", "error", err.Error())
		return
	}

	songRepository := repositories.NewPostgresRepository(pool, slog)
	songService := services.NewSongService(songRepository)
	songHandler := handlers.NewSongHandler(songService)

	docs.SwaggerInfo.BasePath = "/music_library/v1"
	router := gin.Default()
	router.Use(gin.Recovery(), gin.Logger())
	v1 := router.Group("/music_library/v1")
	{
		v1.GET("/songs", songHandler.GetSongs)
		v1.GET("/songs/:id/text", songHandler.GetSongText)
		v1.POST("/songs", songHandler.AddSong)
		v1.PATCH("/songs/:id", songHandler.UpdateSong)
		v1.DELETE("/songs/:id", songHandler.DeleteSong)
	}
	//...
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	err = router.Run(cfg.ServerHost + ":" + cfg.ServerPort)
	if err != nil {
		slog.Error("failed to start server", "err", err.Error(), "port", cfg.ServerPort)
		return
	}
}
