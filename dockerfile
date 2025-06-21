# Use Golang image with Alpine as the base image
FROM golang:1.23-alpine

# Install necessary dependencies: sqlite, GCC, and musl-dev (for cgo)
RUN apk update && apk add --no-cache sqlite sqlite-libs gcc g++ musl-dev

# Set the working directory inside the container
WORKDIR /app

# Add metadata labels
LABEL maintainer="Zakaria.Diouri@outlook.com" \
      version="1.0" \
      description="Go-based forum project" \
      project="Forum Project" \
      Team_members="Zakaria Diouri, Zakaria Kahlaoui, Oussama Atmani, Mohammed-Amine Elayachi, soufiane el walid." \
      created_at="2025-02-18"

# Copy the Go module files and download the dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go project (with cgo enabled for sqlite3 support)
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main .

# Expose the port your app will use
EXPOSE 8090

# Run the Go binary when the container starts
CMD ["./main"]
