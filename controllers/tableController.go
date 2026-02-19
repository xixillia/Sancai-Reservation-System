package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTables menampilkan semua daftar meja
// @Summary Ambil semua meja
// @Description Mengambil daftar lengkap meja beserta statusnya
// @Tags Tables
// @Security BearerAuth
// @Produce json
// @Success 200 {array} structs.Table
// @Failure 500 {object} map[string]string
// @Router /tables [get]
func GetTables(c *gin.Context, DB *sql.DB) {
	var tables []structs.Table
	rows, err := DB.Query("SELECT id, table_number, capacity, status FROM tables")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var table structs.Table
		if err := rows.Scan(&table.ID, &table.TableNumber, &table.Capacity, &table.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		tables = append(tables, table)
	}
	c.JSON(http.StatusOK, tables)
}

// GetTablesByID menampilkan detail satu meja
// @Summary Ambil meja berdasarkan ID
// @Description Mengambil data detail satu meja menggunakan parameter ID
// @Tags Tables
// @Security BearerAuth
// @Produce json
// @Param id path int true "Table ID"
// @Success 200 {object} structs.Table
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tables/{id} [get]
func GetTablesByID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var table structs.Table

	query := "SELECT id, table_number, capacity, status FROM tables WHERE id = $1"
	if err := DB.QueryRow(query, id).Scan(&table.ID, &table.TableNumber, &table.Capacity, &table.Status); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tables not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, table)
}

// CreateTable membuat data meja baru
// @Summary Tambah meja baru
// @Description Menambahkan meja ke sistem (Admin Only)
// @Tags Tables
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param table body structs.Table true "Data Meja"
// @Success 201 {object} structs.Table
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tables [post]
func CreateTable(c *gin.Context, DB *sql.DB) {
	var table structs.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if table.TableNumber == "" || table.Capacity <= 0 || (table.Status != "available" && table.Status != "reserved" && table.Status != "occupied") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table data"})
		return
	}

	//unique number
	var existingID int
	err := DB.QueryRow("SELECT id FROM tables WHERE table_number = $1", table.TableNumber).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Table number already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	query := "INSERT INTO tables (table_number, capacity, status) VALUES ($1, $2, $3) RETURNING id"
	err = DB.QueryRow(query, table.TableNumber, table.Capacity, table.Status).Scan(&table.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, table)
}

// UpdateTable memperbarui data meja
// @Summary Update info meja
// @Description Mengubah nomor meja, kapasitas, atau status
// @Tags Tables
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Table ID"
// @Param table body structs.Table true "Update Data Meja"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tables/{id} [put]
func UpdateTable(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var table structs.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if table.TableNumber == "" || table.Capacity <= 0 || (table.Status != "available" && table.Status != "reserved" && table.Status != "occupied") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table data"})
		return
	}

	query := "UPDATE tables SET table_number = $1, capacity = $2, status = $3 WHERE id = $4"
	result, err := DB.Exec(query, table.TableNumber, table.Capacity, table.Status, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Tables not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tables updated successfully"})
}

// DeleteTable menghapus meja
// @Summary Hapus meja
// @Description Menghapus meja jika tidak ada reservasi aktif di masa depan
// @Tags Tables
// @Security BearerAuth
// @Param id path int true "Table ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /tables/{id} [delete]
func DeleteTable(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")

	// Tidak boleh hapus meja kalau masih ada reservasi future
	var existingID int
	err := DB.QueryRow("SELECT id FROM reservations WHERE table_id = $1 AND reservation_time > NOW()", id).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete table with future reservations"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	query := "DELETE FROM tables WHERE id = $1"
	result, err := DB.Exec(query, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Tables not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tables deleted successfully"})
}

// UpdateTableStatus mengubah status meja saja
// @Summary Update status meja
// @Description Mengubah status (available/reserved/occupied) secara spesifik
// @Tags Tables
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Table ID"
// @Param status body structs.Table true "Status Baru (Cukup isi field status)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /tables/{id}/status [patch]
func UpdateTableStatus(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var status struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE tables SET status = $1 WHERE id = $2"
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Tables not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tables status updated successfully"})
}

// GetAvailableTables mencari meja yang tersedia
// @Summary Cari meja tersedia
// @Description Mencari meja berdasarkan jumlah tamu dan waktu tertentu yang tidak bentrok dengan reservasi lain
// @Tags Tables
// @Security BearerAuth
// @Produce json
// @Param datetime query string true "Format: YYYY-MM-DD HH:MM:SS" example("2026-02-20 19:00:00")
// @Param guests query int true "Jumlah tamu" example(4)
// @Success 200 {array} structs.Table
// @Failure 500 {object} map[string]string
// @Router /tables/available [get]
func GetAvailableTables(c *gin.Context, DB *sql.DB) {
	datetime := c.Query("datetime")
	guests := c.Query("guests")
	var tables []structs.Table
	query := `
		SELECT t.id, t.table_number, t.capacity, t.status
		FROM tables t
		WHERE t.capacity >= $1 AND t.status = 'available' AND t.id NOT IN (
			SELECT r.table_id
			FROM reservations r
			WHERE ABS(EXTRACT(EPOCH FROM (r.reservation_datetime - $2::timestamp)) / 60) < 60
				AND r.status IN ('confirmed', 'pending')
		)
		ORDER BY t.table_number
	`
	rows, err := DB.Query(query, guests, datetime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var table structs.Table
		if err := rows.Scan(&table.ID, &table.TableNumber, &table.Capacity, &table.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tables = append(tables, table)
	}

	c.JSON(http.StatusOK, tables)
}
