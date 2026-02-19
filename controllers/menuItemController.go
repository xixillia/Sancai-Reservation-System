package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMenuItems(c *gin.Context, DB *sql.DB) {
	var menuItems []structs.MenuItem
	rows, err := DB.Query("SELECT id, name, price, available FROM menu_items")
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

func UpdateMenuItemAvailability(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var availability struct {
		Available bool `json:"available"`
	}

	if err := c.ShouldBindJSON(&availability); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE menu_items SET available = $1 WHERE id = $2"
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
