package http

import (
	"database/sql"

	"multifinance-core/internal/handler"
	"multifinance-core/internal/repository"
	"multifinance-core/internal/usecase"

	"github.com/gin-gonic/gin"
)

func NewRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// repository
	authRepo := repository.NewAuthRepo(db)
	consumerRepo := repository.NewConsumerRepo()
	consumerLimitRepo := repository.NewConsumerLimitRepo(db)
	consumerTxRepo := repository.NewConsumerTransactionRepo(db)
	assetRepo := repository.NewAssetRepo(db)

	// usecase
	authUC := usecase.NewAuthUsecase(db, consumerRepo, authRepo)
	assetUC := usecase.NewAssetUsecase(db, assetRepo)
	consumerLimitUC := usecase.NewConsumerLimitUsecase(db, consumerLimitRepo)
	consumerTxUC := usecase.NewConsumerTransactionUsecase(db, assetRepo, consumerLimitRepo, consumerTxRepo)

	// handler
	authHandler := handler.NewAuthHandler(authUC)
	assetHandler := handler.NewAssetHandler(assetUC)
	consumerLimitHandler := handler.NewConsumerLimitHandler(consumerLimitUC)
	consumerTxHandler := handler.NewConsumerTransactionHandler(consumerTxUC)

	// middleware
	authMiddleware := handler.AuthMiddleware(authRepo)

	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		consumers := api.Group("/consumers")
		consumers.Use(authMiddleware)
		{
			consumers.POST("limits/:tenor/use", consumerLimitHandler.Use)
			consumers.POST("transactions", consumerTxHandler.Purchase)
			consumers.GET("transactions", consumerTxHandler.List)
		}

		assets := api.Group("/assets")
		{
			assets.POST("", assetHandler.Create)
			assets.GET("", assetHandler.List)
			assets.GET(":id", assetHandler.Get)
			assets.PUT(":id", assetHandler.Update)
			assets.DELETE(":id", assetHandler.Delete)
		}
	}

	return r
}
