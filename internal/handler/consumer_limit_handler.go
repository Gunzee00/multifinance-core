package handler

import (
    "net/http"
    "strconv"

    "multifinance-core/internal/domain/entity"
    "multifinance-core/internal/usecase"

    "github.com/gin-gonic/gin"
)

type ConsumerLimitHandler struct {
    uc *usecase.ConsumerLimitUsecase
}

func NewConsumerLimitHandler(uc *usecase.ConsumerLimitUsecase) *ConsumerLimitHandler {
    return &ConsumerLimitHandler{uc: uc}
}

type useRequest struct {
    Amount float64 `json:"amount" binding:"required"`
}

func (h *ConsumerLimitHandler) Use(c *gin.Context) {
    // get authenticated user from middleware
    authI, ok := c.Get("auth_user")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }
    authUser := authI.(*entity.AuthUser)
    consumerID := authUser.ConsumerID

    // parse tenor from URL param
    tenorStr := c.Param("tenor")
    t, err := strconv.ParseUint(tenorStr, 10, 8)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenor"})
        return
    }
    tenor := uint8(t)
    if tenor != 1 && tenor != 2 && tenor != 3 && tenor != 6 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenor"})
        return
    }

    var req useRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.uc.IncreaseUsedLimit(c.Request.Context(), consumerID, tenor, req.Amount); err != nil {
        if err.Error() == "used limit exceeds max limit" {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "used limit updated"})
}