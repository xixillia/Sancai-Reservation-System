package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetOrdersByReservationID(c *gin.Context, DB *sql.DB) {
	reservationID := c.Param("id")
	var orders []structs.ReservationOrder

	query := `
		SELECT ro.id, ro.reservation_id, ro.menu_item_id, mi.name, ro.price_at_order, ro.quantity
		FROM reservation_orders ro
		JOIN menu_items mi ON ro.menu_item_id = mi.id
		WHERE ro.reservation_id = $1
	`
	rows, err := DB.Query(query, reservationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var order structs.ReservationOrder
		if err := rows.Scan(&order.ID, &order.ReservationID, &order.MenuItemID, &order.MenuItemName, &order.PriceAtOrder, &order.Quantity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		orders = append(orders, order)
	}
	c.JSON(http.StatusOK, orders)
}

func CreateOrder(c *gin.Context, DB *sql.DB) {
	reservationID := c.Param("id")
	var order structs.ReservationOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO reservation_orders (reservation_id, menu_item_id, quantity) VALUES ($1, $2, $3) RETURNING id"
	err := DB.QueryRow(query, reservationID, order.MenuItemID, order.Quantity).Scan(&order.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func UpdateOrder(c *gin.Context, DB *sql.DB) {
	reservationID := c.Param("id")
	orderID := c.Param("order_id")
	var order structs.ReservationOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE reservation_orders SET quantity = $1 WHERE id = $2 AND reservation_id = $3"
	_, err := DB.Exec(query, order.Quantity, orderID, reservationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
}

func DeleteOrder(c *gin.Context, DB *sql.DB) {
	reservationID := c.Param("id")
	orderID := c.Param("order_id")
	var order structs.ReservationOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "DELETE FROM reservation_orders WHERE id = $1 AND reservation_id = $2"
	_, err := DB.Exec(query, orderID, reservationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}
