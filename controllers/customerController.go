package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// GetCustomers menampilkan semua daftar customer
// @Summary Ambil semua customer
// @Description Mengambil data lengkap semua customer yang terdaftar
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Success 200 {array} structs.Customer
// @Failure 500 {object} map[string]string
// @Router /customers [get]
func GetCustomers(c *gin.Context, DB *sql.DB) {
	var customers []structs.Customer
	rows, err := DB.Query("SELECT id, name, email, phone, created_at FROM customers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var customer structs.Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone, &customer.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		customers = append(customers, customer)
	}

	c.JSON(http.StatusOK, customers)
}

// GetCustomerByID menampilkan detail satu customer
// @Summary Ambil customer berdasarkan ID
// @Description Mengambil data detail satu customer menggunakan parameter ID
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Param id path int true "Customer ID"
// @Success 200 {object} structs.Customer
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customers/{id} [get]
func GetCustomerByID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var customer structs.Customer
	query := "SELECT id, name, email, phone, created_at FROM customers WHERE id = $1"
	if err := DB.QueryRow(query, id).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone, &customer.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, customer)
}

// CreateCustomer menambah customer baru
// @Summary Tambah customer
// @Description Mendaftarkan customer baru dengan validasi email dan nomor telepon unik
// @Tags Customer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param customer body structs.Customer true "Data Customer"
// @Success 201 {object} structs.Customer
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customers [post]
func CreateCustomer(c *gin.Context, DB *sql.DB) {
	var customer structs.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if customer.Name == "" || customer.Email == "" || customer.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, email, and phone are required"})
		return
	}

	//email format
	if !isValidEmail(customer.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	//unique email
	var existingID int
	err := DB.QueryRow("SELECT id FROM customers WHERE email = $1", customer.Email).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	//unique phone
	err = DB.QueryRow("SELECT id FROM customers WHERE phone = $1", customer.Phone).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	query := "INSERT INTO customers (name, email, phone) VALUES ($1, $2, $3) RETURNING id"
	if err := DB.QueryRow(query, customer.Name, customer.Email, customer.Phone).Scan(&customer.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, customer)
}

// UpdateCustomer memperbarui data customer
// @Summary Update data customer
// @Description Mengubah nama, email, atau telepon customer berdasarkan ID
// @Tags Customer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Customer ID"
// @Param customer body structs.Customer true "Update Data Customer"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /customers/{id} [put]
func UpdateCustomer(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var customer structs.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//email format
	if !isValidEmail(customer.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	//unique email
	var existingID int
	err := DB.QueryRow("SELECT id FROM customers WHERE email = $1 AND id != $2", customer.Email, id).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	//unique phone
	err = DB.QueryRow("SELECT id FROM customers WHERE phone = $1 AND id != $2", customer.Phone, id).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	query := "UPDATE customers SET name = $1, email = $2, phone = $3 WHERE id = $4"
	result, err := DB.Exec(query, customer.Name, customer.Email, customer.Phone, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer updated successfully"})
}

// DeleteCustomer menghapus data customer
// @Summary Hapus customer
// @Description Menghapus data customer secara permanen dari database
// @Tags Customer
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /customers/{id} [delete]
func DeleteCustomer(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")

	var count int
	DB.QueryRow("SELECT COUNT(*) FROM reservations WHERE customer_id = $1", id).Scan(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete customer with existing reservations"})
		return
	}

	query := "DELETE FROM customers WHERE id = $1"
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// GetReservationsByCustomerID melihat riwayat reservasi customer
// @Summary Riwayat reservasi customer
// @Description Mengambil semua daftar reservasi yang pernah dibuat oleh customer tertentu
// @Tags Customer
// @Security BearerAuth
// @Produce json
// @Param id path int true "Customer ID"
// @Success 200 {array} structs.Reservation
// @Failure 500 {object} map[string]string
// @Router /customers/{id}/reservations [get]
func GetReservationsByCustomerID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var reservations []structs.Reservation
	query := "SELECT id, customer_id, table_id, reservation_datetime, number_of_guests, status, created_at FROM reservations WHERE customer_id = $1"
	rows, err := DB.Query(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var reservation structs.Reservation
		if err := rows.Scan(&reservation.ID, &reservation.CustomerID, &reservation.TableID, &reservation.ReservationDatetime, &reservation.NumberOfGuests, &reservation.Status, &reservation.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reservations = append(reservations, reservation)
	}
	c.JSON(http.StatusOK, reservations)
}

// Validation email format
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
