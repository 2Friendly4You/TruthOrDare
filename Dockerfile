# Use the official Golang image
FROM golang:1.23.4

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy the source code
COPY . .

# Generate swagger documentation
RUN swag init

# Build the Go application
RUN go build -o main .

# Expose the application port
EXPOSE 8080

# Run the executable
CMD ["./main"]
