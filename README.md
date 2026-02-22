# Sancai Reservation System

Sancai Reservation System adalah RESTful API untuk mengelola reservasi restoran, data pelanggan, meja, menu, dan pesanan menu per reservasi. Sistem ini dibuat berdasarkan studi kasus toko mie Sancai yang sebelumnya mengelola reservasi secara manual.

Project ini dibangun menggunakan Golang sebagai backend service, PostgreSQL sebagai database relasional, dokumentasi API menggunakan Swagger, dan dideploy menggunakan Railway.

---

## Features

* Authentication & Authorization menggunakan JWT
* Manajemen Customer
* Manajemen Meja (Tables)
* Manajemen Menu
* Manajemen Reservasi
* Manajemen Pesanan Menu per Reservasi
* Validasi bisnis (double booking, kapasitas meja, waktu reservasi)
* Dokumentasi API menggunakan Swagger UI

---

## Tech Stack

* Golang (Gin Framework)
* PostgreSQL
* Swagger (OpenAPI)
* JWT Authentication
* Deployment: Railway

---

## Project Structure

```bash
.
├── config
├── controllers
├── models
├── routes
├── middlewares
├── migrations
├── main.go
└── README.md
```

---

## Environment Variables

Buat file `.env` atau set environment variable berikut:

```env
DB_HOST=your_db_host
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
JWT_SECRET=your_jwt_secret
```

---

## Installation & Running Locally

1. Clone repository:

```bash
git clone https://github.com/xixillia/sancai-reservation-system.git
cd sancai-reservation-system
```

2. Install dependencies:

```bash
go mod tidy
```

3. Jalankan aplikasi:

```bash
go run main.go
```

Server akan berjalan di:

```
http://localhost:8080
```

Swagger UI dapat diakses di:

```
http://localhost:8080/swagger/index.html
```

---

## Authentication Flow (JWT)

1. Register user
   POST `/api/register`

2. Login user
   POST `/api/login`
   Response akan mengembalikan JWT token.

3. Akses endpoint protected
   Gunakan header:

```
Authorization: Bearer <JWT_TOKEN>
```

Semua endpoint selain login dan register dilindungi oleh JWT middleware.

---

## API Endpoints

### Auth

| Method | Endpoint      | Description   |
| ------ | ------------- | ------------- |
| POST   | /api/register | Register user |
| POST   | /api/login    | Login user    |

### Customers

| Method | Endpoint                         | Description               |
| ------ | -------------------------------- | ------------------------- |
| GET    | /api/customers                   | Get all customers         |
| GET    | /api/customers/{id}              | Get customer detail       |
| POST   | /api/customers                   | Create new customer       |
| PUT    | /api/customers/{id}              | Update customer           |
| DELETE | /api/customers/{id}              | Delete customer           |
| GET    | /api/customers/{id}/reservations | Get customer reservations |

### Menus

| Method | Endpoint                     | Description              |
| ------ | ---------------------------- | ------------------------ |
| GET    | /api/menus                   | Get all menu items       |
| GET    | /api/menus/{id}              | Get menu item detail     |
| POST   | /api/menus                   | Create menu item         |
| PUT    | /api/menus/{id}              | Update menu item         |
| PATCH  | /api/menus/{id}/availability | Update menu availability |
| DELETE | /api/menus/{id}              | Delete menu item         |

### Tables

| Method | Endpoint                | Description                      |
| ------ | ----------------------- | -------------------------------- |
| GET    | /api/tables             | Get all tables                   |
| GET    | /api/tables/{id}        | Get table detail                 |
| POST   | /api/tables             | Create table                     |
| PUT    | /api/tables/{id}        | Update table                     |
| PATCH  | /api/tables/{id}/status | Update table status              |
| DELETE | /api/tables/{id}        | Delete table                     |
| GET    | /api/tables/available   | Get available tables by datetime |

### Reservations

| Method | Endpoint                      | Description                    |
| ------ | ----------------------------- | ------------------------------ |
| GET    | /api/reservations             | Get all reservations           |
| GET    | /api/reservations/{id}        | Get reservation detail         |
| POST   | /api/reservations             | Create reservation             |
| PUT    | /api/reservations/{id}        | Update reservation             |
| PATCH  | /api/reservations/{id}/status | Update reservation status      |
| DELETE | /api/reservations/{id}        | Delete reservation             |
| GET    | /api/reservations/{id}/total  | Get total price of reservation |

### Reservation Orders

| Method | Endpoint                                 | Description                  |
| ------ | ---------------------------------------- | ---------------------------- |
| GET    | /api/reservations/{id}/orders            | Get orders in reservation    |
| POST   | /api/reservations/{id}/orders            | Add menu item to reservation |
| PUT    | /api/reservations/{id}/orders/{order_id} | Update order quantity        |
| DELETE | /api/reservations/{id}/orders/{order_id} | Delete order item            |

---

## Business Rules & Validations

* Tidak boleh membuat reservasi untuk waktu yang sudah lewat
* Jumlah tamu tidak boleh melebihi kapasitas meja
* Reservasi minimal dibuat 1 jam sebelum waktu yang diinginkan
* Tidak boleh double booking meja pada waktu yang sama
* Menu harus tersedia sebelum bisa dipesan

---

## Deployment

Aplikasi di-deploy menggunakan Railway dengan auto build & deploy dari GitHub.
PostgreSQL dikelola sebagai managed database service di Railway.

---

## Future Enhancements

* Dashboard admin dan staf
* Integrasi pembayaran online
* Pengaturan jam operasional dan hari libur restoran
* Sistem notifikasi otomatis (email/WhatsApp)
