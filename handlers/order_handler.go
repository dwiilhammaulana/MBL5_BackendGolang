package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dwiilhammaulana/gin-firebase-backend/models"
	"github.com/dwiilhammaulana/gin-firebase-backend/services"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{orderService: services.NewOrderService()}
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	order, err := h.orderService.Checkout(userID, &req)
	if err != nil {
		writeOrderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orders, err := h.orderService.GetMyOrders(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil daftar order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func (h *OrderHandler) GetOrderDetail(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID order tidak valid"})
		return
	}

	order, err := h.orderService.GetOrderDetail(userID, uint(orderID))
	if err != nil {
		writeOrderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

func writeOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidPaymentMethod), errors.Is(err, services.ErrCartEmpty):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrProductNotFound), errors.Is(err, services.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrStockNotEnough):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Terjadi kesalahan server"})
	}
}
