package controllers

import (
	"Sancai/structs"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetReservations(c *gin.Context, DB *sql.DB) {
	var reservations []structs.Reservation
	rows, err := DB.Query("SELECT id, customer_id, table_id, reservation_datetime, status FROM reservations")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var reservation structs.Reservation
		if err := rows.Scan(&reservation.ID, &reservation.CustomerID, &reservation.TableID, &reservation.ReservationDatetime, &reservation.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reservations = append(reservations, reservation)
	}
	c.JSON(http.StatusOK, reservations)
}

func GetReservationByID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var reservation structs.Reservation
	query := "SELECT id, customer_id, table_id, reservation_datetime, status FROM reservations WHERE id = $1"
	if err := DB.QueryRow(query, id).Scan(&reservation.ID, &reservation.CustomerID, &reservation.TableID, &reservation.ReservationDatetime, &reservation.Status); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, reservation)
}

func CreateReservation(c *gin.Context, DB *sql.DB) {
	var reservation structs.Reservation
	if err := c.ShouldBindJSON(&reservation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO reservations (customer_id, table_id, reservation_datetime, status) VALUES ($1, $2, $3, $4) RETURNING id"
	err := DB.QueryRow(query, reservation.CustomerID, reservation.TableID, reservation.ReservationDatetime, reservation.Status).Scan(&reservation.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, reservation)
}

func UpdateReservation(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var reservation structs.Reservation
	if err := c.ShouldBindJSON(&reservation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE reservations SET customer_id = $1, table_id = $2, reservation_datetime = $3, status = $4 WHERE id = $5"
	result, err := DB.Exec(query, reservation.CustomerID, reservation.TableID, reservation.ReservationDatetime, reservation.Status, id)
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

func UpdateReservationStatus(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var status struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

func DeleteReservation(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	query := "DELETE FROM reservations WHERE id = $1"
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reservation deleted successfully"})
}

func GetTotalByReservationID(c *gin.Context, DB *sql.DB) {
	id := c.Param("id")
	var total float64
	query := `SELECT SUM(quantity * price_at_order) FROM reservation_orders WHERE reservation_id = $1`
	if err := DB.QueryRow(query, id).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total})
}
