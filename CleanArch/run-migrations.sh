#!/bin/sh
echo "Waiting for MySQL to be ready..."
until nc -z mysql 3306; do
  echo "MySQL is unavailable - sleeping"
  sleep 1
done
echo "MySQL is up - executing migrations"

# Check if table already exists
EXISTING_TABLE=$(mysql -h mysql -u root -proot -N -s -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='orders' AND table_name='orders';" 2>/dev/null || echo "0")

if [ "$EXISTING_TABLE" = "0" ]; then
    echo "Creating orders table..."
    mysql -h mysql -u root -proot -D orders -e "CREATE TABLE orders (id VARCHAR(255) PRIMARY KEY, price FLOAT NOT NULL, tax FLOAT NOT NULL, final_price FLOAT NOT NULL);" 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "Table created successfully!"
    else
        echo "Failed to create table, trying alternative method..."
        mysql -h mysql -u root -proot -D orders < /migrations/1_create_orders_table.up.sql 2>/dev/null
    fi
else
    echo "Table 'orders' already exists, skipping migration."
fi

echo "Migrations completed successfully!"