#!/bin/bash

# Default file name
DEFAULT_FILE="backup.sql"

# Load environment variables from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
else
    echo ".env file not found!"
    exit 1
fi

# Function to export the database
export_db() {
    local file=${1:-$DEFAULT_FILE}
    echo "Exporting database to $file..."
    mysqldump -u $MYSQL_USER -p$MYSQL_PASSWORD -h $MYSQL_HOST -P $MYSQL_PORT $MYSQL_DATABASE > $file
    if [ $? -eq 0 ]; then
        echo "Database exported successfully to $file."
    else
        echo "Failed to export the database."
    fi
}

# Function to import the database
import_db() {
    local file=${1:-$DEFAULT_FILE}
    echo "Importing database from $file..."
    mysql -u $MYSQL_USER -p$MYSQL_PASSWORD -h $MYSQL_HOST -P $MYSQL_PORT $MYSQL_DATABASE < $file
    if [ $? -eq 0 ]; then
        echo "Database imported successfully from $file."
    else
        echo "Failed to import the database."
    fi
}

# Main script logic
if [ $# -eq 0 ]; then
    echo "What do you want to do? (export/import)"
    read action
    echo "Enter file name (default: $DEFAULT_FILE):"
    read file
    file=${file:-$DEFAULT_FILE}
else
    action=$1
    file=${2:-$DEFAULT_FILE}
fi

case $action in
    export)
        export_db $file
        ;;
    import)
        import_db $file
        ;;
    *)
        echo "Invalid action. Use 'export' or 'import'."
        ;;
esac