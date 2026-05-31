package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dwiilhammaulana/gin-firebase-backend/models"
	"github.com/dwiilhammaulana/gin-firebase-backend/services"
	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService *services.CartService
}

func NewCartHandler() *CartHandler {
	return &CartHandler{cartService: services.NewCartService()}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	cart, err := h.cartService.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil keranjang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cart})
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.cartService.AddToCart(userID, &req); err != nil {
		writeCartError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Produk ditambahkan ke keranjang"})
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID item tidak valid"})
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.cartService.UpdateItem(userID, uint(itemID), &req); err != nil {
		writeCartError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keranjang diperbarui"})
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID item tidak valid"})
		return
	}

	if err := h.cartService.RemoveItem(userID, uint(itemID)); err != nil {
		writeCartError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item dihapus dari keranjang"})
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
		return
	}

	if err := h.cartService.ClearCart(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengosongkan keranjang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keranjang dikosongkan"})
}

func writeCartError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrProductNotFound), errors.Is(err, services.ErrCartItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrStockNotEnough):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Terjadi kesalahan server"})
	}
}
