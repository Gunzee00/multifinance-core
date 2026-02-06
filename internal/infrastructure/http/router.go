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

	authRepo := repository.NewAuthRepo(db)
	consumerRepo := repository.NewConsumerRepo()
	consumerLimitRepo := repository.NewConsumerLimitRepo(db)
	consumerTxRepo := repository.NewConsumerTransactionRepo(db)
	assetRepo := repository.NewAssetRepo(db)

	authUC := usecase.NewAuthUsecase(db, consumerRepo, authRepo)
	assetUC := usecase.NewAssetUsecase(db, assetRepo)
	consumerTxUC := usecase.NewConsumerTransactionUsecase(db, assetRepo, consumerLimitRepo, consumerTxRepo)

	authHandler := handler.NewAuthHandler(authUC)
	assetHandler := handler.NewAssetHandler(assetUC)
	consumerTxHandler := handler.NewConsumerTransactionHandler(consumerTxUC)

	authMiddleware := handler.AuthMiddleware(authRepo)

	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		consumers := api.Group("/consumers")
		consumers.Use(authMiddleware)
		{
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
