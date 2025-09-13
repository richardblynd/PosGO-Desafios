CREATE TABLE orders (
    id VARCHAR(255) PRIMARY KEY,
    price FLOAT NOT NULL,
    tax FLOAT NOT NULL,
    final_price FLOAT NOT NULL
);