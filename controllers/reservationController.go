package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetReservations menampilkan semua daftar reservasi
// @Summary Ambil semua reservasi
// @Description Mengambil semua data reservasi yang terdaftar di sistem
// @Tags Reservation
// @Security BearerAuth
// @Produce json
// @Success 200 {array} structs.Reservation
// @Failure 500 {object} map[string]string
// @Router /reservations [get]
func GetReservations(c *gin.Context, DB *sql.DB) {
	var reservations []structs.Reservation
	rows, err := DB.Query("SELECT id, customer_id, table_id, reservation_datetime, status, number_of_guests, created_at FROM reservations")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var reservation structs.Reservation
		if err := rows.Scan(&reservation.ID, &reservation.CustomerID, &reservation.TableID, &reservation.ReservationDatetime, &reservation.Status, &reservation.NumberOfGuests, &reservation.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reservations = append(reservations, reservation)
	}
	c.JSON(http.StatusOK, reservations)
}

// GetReservationByID menampilkan detail satu reservasi
// @Summary Ambil reservasi berdasarkan ID
// @Description Mengambil informasi detail satu reservasi menggunakan ID
// @Tags Reservation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Reservation ID"
// @Success 200 {object} structs.Reservation
// @Failure 404 {object} map[string]string
// @Router /reservations/{id} [get]
func GetReservationByID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var reservation structs.Reservation
	query := "SELECT id, customer_id, table_id, reservation_datetime, status, number_of_guests, created_at FROM reservations WHERE id = $1"
	if err := DB.QueryRow(query, id).Scan(&reservation.ID, &reservation.CustomerID, &reservation.TableID, &reservation.ReservationDatetime, &reservation.Status, &reservation.NumberOfGuests, &reservation.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, reservation)
}

// CreateReservation membuat reservasi baru
// @Summary Buat reservasi baru
// @Description Menambahkan data reservasi baru ke sistem
// @Tags Reservation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param reservation body structs.Reservation true "Data Reservasi"
// @Success 201 {object} structs.Reservation
// @Failure 400 {object} map[string]string
// @Router /reservations [post]
func CreateReservation(c *gin.Context, DB *sql.DB) {
	var reservation structs.Reservation
	if err := c.ShouldBindJSON(&reservation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if reservation.Status == "" {
		reservation.Status = "pending"
	}

	if reservation.CustomerID == 0 || reservation.TableID == 0 || reservation.NumberOfGuests <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CustomerID, TableID, and NumberOfGuests are required"})
		return
	}
	if reservation.ReservationDatetime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ReservationDatetime is required and must be a valid datetime"})
		return
	}

	var count int
	DB.QueryRow("SELECT COUNT(*) FROM customers WHERE id = $1", reservation.CustomerID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID does not exist"})
		return
	}

	DB.QueryRow("SELECT COUNT(*) FROM tables WHERE id = $1", reservation.TableID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Table ID does not exist"})
		return
	}

	// Tidak boleh booking di masa lalu
	if reservation.ReservationDatetime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot book in the past"})
		return
	}

	// Meja yang sama minimal jarak 1 jam
	var existingID int
	DB.QueryRow("SELECT id FROM reservations WHERE table_id = $1 AND ABS(EXTRACT(EPOCH FROM (reservation_datetime - $2::timestamp)) / 60) < 60", reservation.TableID, reservation.ReservationDatetime).Scan(&existingID)
	if existingID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Table is already booked with another reservation"})
		return
	}

	// tables.capacity >= number_of_guests
	var capacity int
	DB.QueryRow("SELECT capacity FROM tables WHERE id = $1", reservation.TableID).Scan(&capacity)
	if reservation.NumberOfGuests > capacity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number of guests exceeds table capacity"})
		return
	}

	// Status hanya boleh: pending, confirmed, cancelled, completed, kosong (default pending)
	if reservation.Status != "pending" && reservation.Status != "confirmed" && reservation.Status != "cancelled" && reservation.Status != "completed" && reservation.Status != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'pending', 'confirmed', 'cancelled', or 'completed'"})
		return
	}

	// Customer tidak boleh punya 2 reservasi aktif di jam yang sama
	var activeID int
	DB.QueryRow("SELECT id FROM reservations WHERE customer_id = $1 AND ABS(EXTRACT(EPOCH FROM (reservation_datetime - $2::timestamp)) / 60) < 60 AND status IN ('pending', 'confirmed')", reservation.CustomerID, reservation.ReservationDatetime).Scan(&activeID)
	if activeID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer already has an active reservation at the same time"})
		return
	}

	query := "INSERT INTO reservations (customer_id, table_id, reservation_datetime, status, number_of_guests) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := DB.QueryRow(query, reservation.CustomerID, reservation.TableID, reservation.ReservationDatetime, reservation.Status, reservation.NumberOfGuests).Scan(&reservation.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, reservation)
}

// UpdateReservation memperbarui data reservasi
// @Summary Update data reservasi
// @Description Mengubah informasi customer, meja, waktu, atau status reservasi
// @Tags Reservation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param reservation body structs.Reservation true "Update Data Reservasi"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /reservations/{id} [put]
func UpdateReservation(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var reservation structs.Reservation
	if err := c.ShouldBindJSON(&reservation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if reservation.CustomerID == 0 || reservation.TableID == 0 || reservation.NumberOfGuests <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CustomerID, TableID, and NumberOfGuests are required"})
		return
	}
	if reservation.ReservationDatetime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ReservationDatetime is required and must be a valid datetime"})
		return
	}

	var count int
	DB.QueryRow("SELECT COUNT(*) FROM customers WHERE id = $1", reservation.CustomerID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID does not exist"})
		return
	}

	DB.QueryRow("SELECT COUNT(*) FROM tables WHERE id = $1", reservation.TableID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Table ID does not exist"})
		return
	}

	// Tidak boleh booking di masa lalu
	if reservation.ReservationDatetime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot book in the past"})
		return
	}

	// Meja yang sama minimal jarak 1 jam
	var existingID int
	DB.QueryRow("SELECT id FROM reservations WHERE table_id = $1 AND id != $2 AND ABS(EXTRACT(EPOCH FROM (reservation_datetime - $3::timestamp)) / 60) < 60", reservation.TableID, id, reservation.ReservationDatetime).Scan(&existingID)
	if existingID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Table is already booked with another reservation"})
		return
	}

	// tables.capacity >= number_of_guests
	var capacity int
	DB.QueryRow("SELECT capacity FROM tables WHERE id = $1", reservation.TableID).Scan(&capacity)
	if reservation.NumberOfGuests > capacity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number of guests exceeds table capacity"})
		return
	}

	// Status hanya boleh: pending, confirmed, cancelled, completed, kosong (default pending)
	if reservation.Status != "pending" && reservation.Status != "confirmed" && reservation.Status != "cancelled" && reservation.Status != "completed" && reservation.Status != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'pending', 'confirmed', 'cancelled', or 'completed'"})
		return
	}

	// Customer tidak boleh punya 2 reservasi aktif di jam yang sama
	var activeID int
	DB.QueryRow("SELECT id FROM reservations WHERE customer_id = $1 AND id != $2 AND ABS(EXTRACT(EPOCH FROM (reservation_datetime - $3::timestamp)) / 60) < 60 AND status IN ('pending', 'confirmed')", reservation.CustomerID, id, reservation.ReservationDatetime).Scan(&activeID)
	if activeID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer already has an active reservation at the same time"})
		return
	}

	query := "UPDATE reservations SET customer_id = $1, table_id = $2, reservation_datetime = $3 WHERE id = $4"
	result, err := DB.Exec(query, reservation.CustomerID, reservation.TableID, reservation.ReservationDatetime, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reservation updated successfully"})
}

// UpdateReservationStatus mengubah status reservasi
// @Summary Update status reservasi
// @Description Mengubah status (pending/confirmed/cancelled/completed)
// @Tags Reservation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param status body structs.Reservation true "Cukup isi field status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /reservations/{id}/status [patch]
func UpdateReservationStatus(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var status struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Status hanya boleh: pending, confirmed, cancelled, completed, kosong (default pending)
	if status.Status != "pending" && status.Status != "confirmed" && status.Status != "cancelled" && status.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'pending', 'confirmed', 'cancelled', or 'completed'"})
		return
	}

	query := "UPDATE reservations SET status = $1 WHERE id = $2"
	result, err := DB.Exec(query, status.Status, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation status updated successfully"})
}

// DeleteReservation menghapus data reservasi
// @Summary Hapus reservasi
// @Description Menghapus record reservasi dari database
// @Tags Reservation
// @Security BearerAuth
// @Param id path int true "Reservation ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /reservations/{id} [delete]
func DeleteReservation(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")

	// cek apakah reservasi ada
	var existingID int
	DB.QueryRow("SELECT id FROM reservations WHERE id = $1", id).Scan(&existingID)
	if existingID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
		return
	}

	// status tidak boleh completed / confirmed
	var status string
	DB.QueryRow("SELECT status FROM reservations WHERE id = $1", id).Scan(&status)
	if status == "completed" || status == "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete a reservation that is completed or confirmed"})
		return
	}

	// Hapus juga reservation_orders terkait
	_, err := DB.Exec("DELETE FROM reservation_orders WHERE reservation_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	query := "DELETE FROM reservations WHERE id = $1"
	_, err = DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation deleted successfully"})
}

// GetTotalByReservationID menghitung total biaya pesanan
// @Summary Ambil total harga reservasi
// @Description Menghitung jumlah (quantity * price) dari semua item yang dipesan dalam satu reservasi
// @Tags Reservation
// @Security BearerAuth
// @Produce json
// @Param id path int true "Reservation ID"
// @Success 200 {object} map[string]float64 "Contoh: {"total": 150000}"
// @Failure 500 {object} map[string]string
// @Router /reservations/{id}/total [get]
func GetTotalByReservationID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var total float64
	query := `SELECT COALESCE(SUM(quantity * price_at_order), 0) FROM reservation_orders WHERE reservation_id = $1`
	if err := DB.QueryRow(query, id).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total})
}
