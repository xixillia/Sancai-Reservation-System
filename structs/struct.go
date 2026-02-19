package structs

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Customer struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type Table struct {
	ID          int    `json:"id"`
	TableNumber string `json:"table_number"`
	Capacity    int    `json:"capacity"`
	Status      string `json:"status"`
}

type Reservation struct {
	ID                  int       `json:"id"`
	CustomerID          int       `json:"customer_id"`
	TableID             int       `json:"table_id"`
	ReservationDatetime string    `json:"reservation_datetime"`
	NumberOfGuests      int       `json:"number_of_guests"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
}

type MenuItem struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	IsAvailable bool    `json:"is_available"`
}

type ReservationOrder struct {
	ID            int     `json:"id"`
	ReservationID int     `json:"reservation_id"`
	MenuItemID    int     `json:"menu_item_id"`
	MenuItemName  string  `json:"menu_item_name"`
	Quantity      int     `json:"quantity"`
	PriceAtOrder  float64 `json:"price_at_order"`
}
