-- +migrate Up

-- customers
CREATE TABLE customers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(150) NOT NULL,
  email VARCHAR(150),
  phone VARCHAR(50),
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- tables (meja restoran)
CREATE TABLE tables (
  id BIGSERIAL PRIMARY KEY,
  table_number INT NOT NULL,
  capacity INT NOT NULL,
  status VARCHAR(50) NOT NULL
);

-- reservations
CREATE TABLE reservations (
  id BIGSERIAL PRIMARY KEY,
  customer_id BIGINT NOT NULL,
  table_id BIGINT NOT NULL,
  reservation_datetime TIMESTAMP NOT NULL,
  number_of_guests INT NOT NULL,
  status VARCHAR(50) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),

  CONSTRAINT fk_reservations_customer
    FOREIGN KEY (customer_id)
    REFERENCES customers(id)
    ON DELETE RESTRICT,

  CONSTRAINT fk_reservations_table
    FOREIGN KEY (table_id)
    REFERENCES tables(id)
    ON DELETE RESTRICT
);

--  menu_items
CREATE TABLE menu_items (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(150) NOT NULL,
  price NUMERIC(12,2) NOT NULL,
  category VARCHAR(100),
  is_available BOOLEAN NOT NULL DEFAULT TRUE
);

-- 5️⃣ reservation_orders
CREATE TABLE reservation_orders (
  id BIGSERIAL PRIMARY KEY,
  reservation_id BIGINT NOT NULL,
  menu_item_id BIGINT NOT NULL,
  quantity INT NOT NULL,
  price_at_order NUMERIC(12,2) NOT NULL,

  CONSTRAINT fk_orders_reservation
    FOREIGN KEY (reservation_id)
    REFERENCES reservations(id)
    ON DELETE RESTRICT,

  CONSTRAINT fk_orders_menu_item
    FOREIGN KEY (menu_item_id)
    REFERENCES menu_items(id)
    ON DELETE RESTRICT
);
