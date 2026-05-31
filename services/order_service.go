package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/dwiilhammaulana/gin-firebase-backend/config"
	"github.com/dwiilhammaulana/gin-firebase-backend/models"
	"github.com/dwiilhammaulana/gin-firebase-backend/repositories"
	"gorm.io/gorm"
)

var (
	ErrCartEmpty            = errors.New("keranjang kosong")
	ErrInvalidPaymentMethod = errors.New("metode pembayaran tidak valid")
	ErrOrderNotFound        = errors.New("order tidak ditemukan")
)

type OrderService struct {
	cartRepo  *repositories.CartRepository
	orderRepo *repositories.OrderRepository
}

func NewOrderService() *OrderService {
	return &OrderService{
		cartRepo:  repositories.NewCartRepository(),
		orderRepo: repositories.NewOrderRepository(),
	}
}

func (s *OrderService) Checkout(userID uint, req *models.CheckoutRequest) (*models.OrderResponse, error) {
	if !isAllowedPaymentMethod(req.PaymentMethod) {
		return nil, ErrInvalidPaymentMethod
	}

	var createdOrder *models.Order
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		cartTx := s.cartRepo.WithTransaction(tx)
		orderTx := s.orderRepo.WithTransaction(tx)

		cartItems, err := cartTx.FindByUserID(userID)
		if err != nil {
			return err
		}
		if len(cartItems) == 0 {
			return ErrCartEmpty
		}

		var total float64
		orderItems := make([]models.OrderItem, 0, len(cartItems))
		for _, item := range cartItems {
			if !item.Product.IsActive {
				return ErrProductNotFound
			}
			if item.Product.Stock < item.Quantity {
				return ErrStockNotEnough
			}

			subtotal := item.Product.Price * float64(item.Quantity)
			total += subtotal
			orderItems = append(orderItems, models.OrderItem{
				ProductID:   item.ProductID,
				ProductName: item.Product.Name,
				Price:       item.Product.Price,
				Quantity:    item.Quantity,
				Subtotal:    subtotal,
			})
		}

		vaNumber, gopayDeeplink := generatePaymentData(req.PaymentMethod, userID)
		order := &models.Order{
			UserID:          userID,
			TotalAmount:     total,
			Status:          "pending",
			ShippingAddress: req.ShippingAddress,
			Notes:           req.Notes,
			PaymentMethod:   req.PaymentMethod,
			PaymentStatus:   "pending",
			VANumber:        vaNumber,
			GopayDeeplink:   gopayDeeplink,
		}

		if err := orderTx.Create(order); err != nil {
			return err
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
			rowsAffected, err := orderTx.DecreaseProductStock(orderItems[i].ProductID, orderItems[i].Quantity)
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				return ErrStockNotEnough
			}
		}

		if err := orderTx.CreateOrderItems(orderItems); err != nil {
			return err
		}

		payment := &models.Payment{
			OrderID:       order.ID,
			PaymentMethod: req.PaymentMethod,
			PaymentStatus: "pending",
			Amount:        total,
			VANumber:      vaNumber,
			GopayDeeplink: gopayDeeplink,
		}
		if err := orderTx.CreatePayment(payment); err != nil {
			return err
		}

		if err := cartTx.ClearByUserID(userID); err != nil {
			return err
		}

		order.Items = orderItems
		createdOrder = order
		return nil
	})
	if err != nil {
		return nil, err
	}

	response := buildOrderResponse(*createdOrder)
	return &response, nil
}

func (s *OrderService) GetMyOrders(userID uint, page, limit int) (*models.OrdersListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	orders, total, err := s.orderRepo.FindByUserID(userID, page, limit)
	if err != nil {
		return nil, err
	}

	items := make([]models.OrderResponse, 0, len(orders))
	for _, order := range orders {
		items = append(items, buildOrderResponse(order))
	}

	return &models.OrdersListResponse{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *OrderService) GetOrderDetail(userID, orderID uint) (*models.OrderResponse, error) {
	order, err := s.orderRepo.FindByIDAndUserID(orderID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	response := buildOrderResponse(*order)
	return &response, nil
}

func isAllowedPaymentMethod(method string) bool {
	return method == "gopay" || method == "bank_transfer" || method == "virtual_account"
}

func generatePaymentData(method string, userID uint) (*string, *string) {
	now := time.Now().Unix()
	switch method {
	case "virtual_account":
		value := fmt.Sprintf("8808%d%d", userID, now)
		return &value, nil
	case "gopay":
		value := fmt.Sprintf("gopay://payment/order/%d/%d", userID, now)
		return nil, &value
	default:
		return nil, nil
	}
}

func buildOrderResponse(order models.Order) models.OrderResponse {
	items := make([]models.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, models.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
		})
	}

	return models.OrderResponse{
		ID:              order.ID,
		TotalAmount:     order.TotalAmount,
		Status:          order.Status,
		ShippingAddress: order.ShippingAddress,
		Notes:           order.Notes,
		PaymentMethod:   order.PaymentMethod,
		PaymentStatus:   order.PaymentStatus,
		VANumber:        order.VANumber,
		GopayDeeplink:   order.GopayDeeplink,
		PaidAt:          order.PaidAt,
		Items:           items,
		CreatedAt:       order.CreatedAt,
	}
}
