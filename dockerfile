FROM golang:1.23-alpine

# Set up alternative Alpine package repositories
RUN echo "http://dl-4.alpinelinux.org/alpine/v3.22/main" > /etc/apk/repositories && \
    echo "http://dl-4.alpinelinux.org/alpine/v3.22/community" >> /etc/apk/repositories

# Debug network connectivity
RUN apk add --no-cache curl && \
    curl -v http://dl-4.alpinelinux.org/alpine/v3.22/main/x86_64/APKINDEX.tar.gz || echo "Failed to fetch APKINDEX"

# Install dependencies
RUN apk update && apk add --no-cache sqlite sqlite-libs gcc g++ musl-dev

# Set up working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o forum

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./forum"]