package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// Order entity for SQL Server storage - demonstrates transactional data
type Order struct {
	ID              uuid.UUID   `json:"id"`
	CustomerID      uuid.UUID   `json:"customerId"`
	OrderDate       time.Time   `json:"orderDate"`
	Status          OrderStatus `json:"status"`
	TotalAmount     float64     `json:"totalAmount"`
	Currency        string      `json:"currency"`
	ShippingAddress string      `json:"shippingAddress"`
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"productId"`
	ProductName string    `json:"productName"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
}

// NewOrder creates a new order with calculated total
func NewOrder(customerID uuid.UUID, shippingAddress string, items []OrderItem) *Order {
	now := time.Now().UTC()

	// Assign IDs to items
	for i := range items {
		if items[i].ID == uuid.Nil {
			items[i].ID = uuid.New()
		}
	}

	order := &Order{
		ID:              uuid.New(),
		CustomerID:      customerID,
		OrderDate:       now,
		Status:          OrderStatusPending,
		Currency:        "USD",
		ShippingAddress: shippingAddress,
		Items:           items,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Calculate total
	order.TotalAmount = order.CalculateTotal()

	return order
}

// CalculateTotal calculates the total amount from items
func (o *Order) CalculateTotal() float64 {
	var total float64
	for _, item := range o.Items {
		total += float64(item.Quantity) * item.Price
	}
	return total
}

// UpdateStatus updates the order status with validation
func (o *Order) UpdateStatus(newStatus OrderStatus) error {
	if !o.CanTransitionTo(newStatus) {
		return &InvalidStatusTransitionError{
			From: o.Status,
			To:   newStatus,
		}
	}
	o.Status = newStatus
	o.UpdatedAt = time.Now().UTC()
	return nil
}

// CanTransitionTo checks if a status transition is valid
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	switch o.Status {
	case OrderStatusPending:
		return newStatus == OrderStatusProcessing || newStatus == OrderStatusCancelled
	case OrderStatusProcessing:
		return newStatus == OrderStatusShipped || newStatus == OrderStatusCancelled
	case OrderStatusShipped:
		return newStatus == OrderStatusDelivered
	case OrderStatusDelivered, OrderStatusCancelled:
		return false
	default:
		return false
	}
}

// InvalidStatusTransitionError represents an invalid status transition
type InvalidStatusTransitionError struct {
	From OrderStatus
	To   OrderStatus
}

func (e *InvalidStatusTransitionError) Error() string {
	return "cannot transition from " + string(e.From) + " to " + string(e.To)
}

// GetTotal returns the item total (quantity × price)
func (i *OrderItem) GetTotal() float64 {
	return float64(i.Quantity) * i.Price
}

// NewOrderItem creates a new order item
func NewOrderItem(productID uuid.UUID, productName string, quantity int, price float64) OrderItem {
	return OrderItem{
		ID:          uuid.New(),
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		Price:       price,
	}
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID      uuid.UUID              `json:"customerId"`
	ShippingAddress string                 `json:"shippingAddress"`
	Items           []CreateOrderItemInput `json:"items"`
}

// CreateOrderItemInput represents input for creating an order item
type CreateOrderItemInput struct {
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
}
