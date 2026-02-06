package handler

import (
	"net/http"

	"multifinance-core/internal/domain/entity"
	"multifinance-core/internal/usecase"

	"github.com/gin-gonic/gin"
)

type ConsumerTransactionHandler struct {
	uc *usecase.ConsumerTransactionUsecase
}

func NewConsumerTransactionHandler(uc *usecase.ConsumerTransactionUsecase) *ConsumerTransactionHandler {
	return &ConsumerTransactionHandler{uc: uc}
}

type purchaseRequest struct {
	AssetID uint64 `json:"asset_id" binding:"required"`
	Tenor   uint8  `json:"tenor" binding:"required"`
}

func (h *ConsumerTransactionHandler) Purchase(c *gin.Context) {
	authI, ok := c.Get("auth_user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authUser := authI.(*entity.AuthUser)
	consumerID := authUser.ConsumerID

	var req purchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tr, err := h.uc.Purchase(c.Request.Context(), consumerID, req.AssetID, req.Tenor)
	if err != nil {
		if err == usecase.ErrInvalidTenor {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == usecase.ErrInsufficientLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient limit"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "purchase success", "transaction": tr})
}

func (h *ConsumerTransactionHandler) List(c *gin.Context) {
	authI, ok := c.Get("auth_user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authUser := authI.(*entity.AuthUser)
	consumerID := authUser.ConsumerID

	list, err := h.uc.ListByConsumer(c.Request.Context(), consumerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": list})
}
