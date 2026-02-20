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
// @Router /menus [get]
func GetMenuItems(c *gin.Context, DB *sql.DB) {
	var menuItems []structs.MenuItem
	rows, err := DB.Query("SELECT id, name, price, category, is_available FROM menu_items")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var menuItem structs.MenuItem
		if err := rows.Scan(&menuItem.ID, &menuItem.Name, &menuItem.Price, &menuItem.Category, &menuItem.IsAvailable); err != nil {
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
// @Router /menus/{id} [get]
func GetMenuItemByID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var menuItem structs.MenuItem
	query := "SELECT id, name, price, category, is_available FROM menu_items WHERE id = $1"
	if err := DB.QueryRow(query, id).Scan(&menuItem.ID, &menuItem.Name, &menuItem.Price, &menuItem.Category, &menuItem.IsAvailable); err != nil {
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
// @Router /menus [post]
func CreateMenuItem(c *gin.Context, DB *sql.DB) {
	var menuItem structs.MenuItem
	if err := c.ShouldBindJSON(&menuItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if menuItem.Name == "" || menuItem.Price <= 0 || menuItem.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, price, and category are required and price must be greater than 0"})
		return
	}

	if menuItem.Category != "main course" && menuItem.Category != "drink" && menuItem.Category != "dessert" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category must be 'main course', 'drink', or 'dessert'"})
		return
	}

	query := "INSERT INTO menu_items (name, price, category, is_available) VALUES ($1, $2, $3, $4) RETURNING id"
	if err := DB.QueryRow(query, menuItem.Name, menuItem.Price, menuItem.Category, menuItem.IsAvailable).Scan(&menuItem.ID); err != nil {
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
// @Router /menus/{id} [put]
func UpdateMenuItem(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var menuItem structs.MenuItem
	if err := c.ShouldBindJSON(&menuItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if menuItem.Name == "" || menuItem.Price <= 0 || menuItem.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, price, and category are required and price must be greater than 0"})
		return
	}

	if menuItem.Category != "main course" && menuItem.Category != "drink" && menuItem.Category != "dessert" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category must be 'main course', 'drink', or 'dessert'"})
		return
	}

	query := "UPDATE menu_items SET name = $1, price = $2, category = $3, is_available = $4 WHERE id = $5"
	result, err := DB.Exec(query, menuItem.Name, menuItem.Price, menuItem.Category, menuItem.IsAvailable, id)
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
// @Router /menus/{id} [delete]
func DeleteMenuItem(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
		
	// Tidak boleh hapus menu kalau masih ada di reservation_orders
	var existingID int
	err := DB.QueryRow("SELECT id FROM reservation_orders WHERE menu_item_id = $1", id).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Menu item cannot be deleted because it is still part of existing orders"})
		}
		return
	}

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
// @Router /menus/{id}/availability [patch]
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
