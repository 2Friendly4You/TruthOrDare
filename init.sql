CREATE DATABASE IF NOT EXISTS itemsdb;

USE itemsdb;

CREATE TABLE IF NOT EXISTS items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

INSERT INTO items (name, price) VALUES
    ('Item 1', 10.50),
    ('Item 2', 15.75),
    ('Item 3', 7.30);
