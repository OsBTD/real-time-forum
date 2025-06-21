#!/bin/bash

# Set variables
IMAGE_NAME="forum-image"
CONTAINER_NAME="forum-container1"
PORT=8080

# Step 1: Build the Docker image
echo "Building Docker image..."
docker build -t $IMAGE_NAME .

# Step 2: Check if a container with the same name is already running and remove it
echo "Stopping and removing any existing container named $CONTAINER_NAME..."
docker rm -f $CONTAINER_NAME || true  # Ignore if container doesn't exist

# Step 3: Run the Docker container
echo "Running Docker container..."
docker run -d -p $PORT:$PORT --name $CONTAINER_NAME $IMAGE_NAME

# Step 4: Show running containers
docker ps
