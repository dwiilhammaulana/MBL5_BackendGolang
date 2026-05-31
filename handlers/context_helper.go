package handlers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getUserIDFromContext(c *gin.Context) (uint, error) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user_id tidak ditemukan")
	}

	switch v := value.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case int64:
		return uint(v), nil
	case float64:
		return uint(v), nil
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		return uint(id), err
	default:
		return 0, errors.New("user_id tidak valid")
	}
}
