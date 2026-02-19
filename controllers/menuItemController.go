package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMenuItems menampilkan semua daftar menu
// @Summary Ambil semua menu
// @Description Mengambil daftar lengkap item menu makanan dan minuman
// @Tags Menu
// @Security BearerAuth
// @Produce json
// @Success 200 {array} structs.MenuItem
// @Failure 500 {object} map[string]string
// @Router /menu [get]
func GetMenuItems(c *gin.Context, DB *sql.DB) {
	var menuItems []structs.MenuItem
	rows, err := DB.Query("SELECT id, name, price, is_available FROM menu_items")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var menuItem structs.MenuItem
		if err := rows.Scan(&menuItem.ID, &menuItem.Name, &menuItem.Price, &menuItem.IsAvailable); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		menuItems = append(menuItems, menuItem)
	}

	c.JSON(http.StatusOK, menuItems)
}

// GetMenuItemByID menampilkan detail satu menu
// @Summary Ambil menu berdasarkan ID
// @Description Mengambil data detail satu item menu menggunakan parameter ID
// @Tags Menu
// @Security BearerAuth
// @Produce json
// @Param id path int true "Menu Item ID"
// @Success 200 {object} structs.MenuItem
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /menu/{id} [get]
func GetMenuItemByID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var menuItem structs.MenuItem
	query := "SELECT id, name, price, is_available FROM menu_items WHERE id = $1"
	if err := DB.QueryRow(query, id).Scan(&menuItem.ID, &menuItem.Name, &menuItem.Price, &menuItem.IsAvailable); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, menuItem)
}

// CreateMenuItem menambah item menu baru
// @Summary Tambah menu baru
// @Description Menambahkan item menu baru ke dalam sistem (Admin Only)
// @Tags Menu
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param menu body structs.MenuItem true "Data Menu Item"
// @Success 201 {object} structs.MenuItem
// @Failure 400 {object} map[string]string
// @Router /menu [post]
func CreateMenuItem(c *gin.Context, DB *sql.DB) {
	var menuItem structs.MenuItem
	if err := c.ShouldBindJSON(&menuItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO menu_items (name, price, is_available) VALUES ($1, $2, $3) RETURNING id"
	if err := DB.QueryRow(query, menuItem.Name, menuItem.Price, menuItem.IsAvailable).Scan(&menuItem.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, menuItem)
}

// UpdateMenuItem memperbarui data menu
// @Summary Update data menu
// @Description Mengubah nama, harga, atau status ketersediaan menu
// @Tags Menu
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Param menu body structs.MenuItem true "Update Data Menu"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /menu/{id} [put]
func UpdateMenuItem(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var menuItem structs.MenuItem
	if err := c.ShouldBindJSON(&menuItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE menu_items SET name = $1, price = $2, is_available = $3 WHERE id = $4"
	result, err := DB.Exec(query, menuItem.Name, menuItem.Price, menuItem.IsAvailable, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Menu item updated successfully"})
}

// DeleteMenuItem menghapus item menu
// @Summary Hapus menu
// @Description Menghapus item menu dari database berdasarkan ID
// @Tags Menu
// @Security BearerAuth
// @Param id path int true "Menu Item ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /menu/{id} [delete]
func DeleteMenuItem(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	query := "DELETE FROM menu_items WHERE id = $1"
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Menu item deleted successfully"})
}

// UpdateMenuItemAvailability mengubah status ketersediaan menu
// @Summary Update ketersediaan menu
// @Description Mengubah status ketersediaan (available true/false) secara cepat
// @Tags Menu
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Param availability body structs.MenuItem true "Cukup isi field is_available"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /menu/{id}/availability [patch]
func UpdateMenuItemAvailability(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var availability struct {
		Available bool `json:"available"`
	}

	if err := c.ShouldBindJSON(&availability); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE menu_items SET is_available = $1 WHERE id = $2"
	result, err := DB.Exec(query, availability.Available, id)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Menu item availability updated successfully"})
}
