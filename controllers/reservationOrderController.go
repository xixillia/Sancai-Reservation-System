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

	query := "INSERT INTO reservation_orders (reservation_id, menu_item_id, quantity) VALUES ($1, $2, $3) RETURNING id"
	err := DB.QueryRow(query, reservationID, order.MenuItemID, order.Quantity).Scan(&order.ID)
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

	query := "UPDATE reservation_orders SET quantity = $1 WHERE id = $2 AND reservation_id = $3"
	_, err := DB.Exec(query, order.Quantity, orderID, reservationID)
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
