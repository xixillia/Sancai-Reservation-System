package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetOrdersByReservationID menampilkan semua item yang dipesan dalam satu reservasi
// @Summary Ambil daftar pesanan per reservasi
// @Description Mengambil semua item menu yang dipesan beserta kuantitas dan harga saat dipesan
// @Tags ReservationOrder
// @Security BearerAuth
// @Produce json
// @Param id path int true "Reservation ID"
// @Success 200 {array} structs.ReservationOrder
// @Failure 500 {object} map[string]string
// @Router /reservations/{id}/orders [get]
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

// CreateOrder menambahkan item menu ke dalam reservasi
// @Summary Tambah pesanan baru
// @Description Menambahkan item menu (menu_item_id) dan jumlahnya ke dalam reservasi tertentu
// @Tags ReservationOrder
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param order body structs.ReservationOrder true "Data Pesanan Item"
// @Success 201 {object} structs.ReservationOrder
// @Failure 400 {object} map[string]string
// @Router /reservations/{id}/orders [post]
func CreateOrder(c *gin.Context, DB *sql.DB) {
	reservationID := c.Param("id")
	var order structs.ReservationOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.MenuItemID == 0 || order.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MenuItemID and Quantity are required"})
		return
	}

	// reservation_id harus ada di reservations
	var count int
	DB.QueryRow("SELECT count(*) FROM reservations WHERE id = $1", reservationID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation ID does not exist"})
		return
	}

	// menu_item_id harus ada di menu_items
	DB.QueryRow("SELECT count(*) FROM menu_items WHERE id = $1 AND is_available = true", order.MenuItemID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Menu does not exist or is not available"})
		return
	}

	// Cek status reservasi
	var reservationStatus string
	err := DB.QueryRow("SELECT status FROM reservations WHERE id = $1", reservationID).Scan(&reservationStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reservation status"})
		return
	}
	if reservationStatus == "cancelled" || reservationStatus == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add order to a cancelled or completed reservation"})
		return
	}

	//get current price_at_order
	var priceAtOrder float64
	err = DB.QueryRow("SELECT price FROM menu_items WHERE id = $1", order.MenuItemID).Scan(&priceAtOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get menu item price"})
		return
	}
	order.PriceAtOrder = priceAtOrder

	//get menu name
	var menuItemName string
	err = DB.QueryRow("SELECT name FROM menu_items WHERE id = $1", order.MenuItemID).Scan(&menuItemName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get menu item name"})
		return
	}
	order.MenuItemName = menuItemName

	query := "INSERT INTO reservation_orders (reservation_id, menu_item_id, quantity, price_at_order) VALUES ($1, $2, $3, $4) RETURNING id"
	err = DB.QueryRow(query, reservationID, order.MenuItemID, order.Quantity, order.PriceAtOrder).Scan(&order.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// UpdateOrder mengubah kuantitas pesanan
// @Summary Update kuantitas pesanan
// @Description Mengubah jumlah pesanan berdasarkan Order ID dan Reservation ID
// @Tags ReservationOrder
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param order_id path int true "Order ID"
// @Param order body structs.ReservationOrder true "Update Kuantitas (cukup field quantity)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /reservations/{id}/orders/{order_id} [put]
func UpdateOrder(c *gin.Context, DB *sql.DB) {
	reservationID := c.Param("id")
	orderID := c.Param("order_id")
	var order structs.ReservationOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity is required"})
		return
	}

	// order_id harus ada di reservation_orders dengan reservation_id yang sesuai
	var count int
	DB.QueryRow("SELECT count(*) FROM reservation_orders WHERE id = $1 AND reservation_id = $2", orderID, reservationID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID does not exist for the given Reservation ID"})
		return
	}

	// Tidak boleh update order kalau reservasi sudah completed
	var reservationStatus string
	err := DB.QueryRow("SELECT status FROM reservations WHERE id = $1", reservationID).Scan(&reservationStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reservation status"})
		return
	}
	if reservationStatus == "completed" || reservationStatus == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update order from a completed or cancelled reservation"})
		return
	}

	query := "UPDATE reservation_orders SET quantity = $1 WHERE id = $2 AND reservation_id = $3"
	_, err = DB.Exec(query, order.Quantity, orderID, reservationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
}

// DeleteOrder menghapus item dari pesanan
// @Summary Hapus pesanan item
// @Description Menghapus satu item menu yang sudah dipesan dari daftar reservasi
// @Tags ReservationOrder
// @Security BearerAuth
// @Param id path int true "Reservation ID"
// @Param order_id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Router /reservations/{id}/orders/{order_id} [delete]
func DeleteOrder(c *gin.Context, DB *sql.DB) {
	reservationID := c.Param("id")
	orderID := c.Param("order_id")

	// Tidak boleh hapus order kalau reservasi sudah completed
	var reservationStatus string
	err := DB.QueryRow("SELECT status FROM reservations WHERE id = $1", reservationID).Scan(&reservationStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reservation status"})
		return
	}
	if reservationStatus == "completed" || reservationStatus == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete order from a completed or cancelled reservation"})
		return
	}

	query := "DELETE FROM reservation_orders WHERE id = $1 AND reservation_id = $2"
	result, err := DB.Exec(query, orderID, reservationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation Order not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reservation Order deleted successfully"})
}
