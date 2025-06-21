# FORUM Project

**Leader**: Zakaria Diouri
**Members**: Zakaria Kahlaoui, Oussama Atmani, Mohammed-Amine Elayachi, soufiane el walid.

## Objectives

The objective of this project is to create a **web forum** that allows:

- Communication between users.
- Associating categories to posts.
- Liking and disliking posts and comments.
- Filtering posts by category, creation date, and user likes.

## Features

### SQLite
- **SQLite** is used for storing forum data, including users, posts, comments, and likes.
- It is a lightweight, embedded database that allows us to handle the data efficiently.

### Authentication
- Users can register by providing their credentials (email, username, and password).
- Passwords are stored securely with encryption (Bonus Task).
- Users can log in using cookies that store session information, and each session will have an expiration date.
- **UUID** (Bonus Task) may be used to uniquely identify sessions.

### User Registration
- Users must provide:
  - **Email** (checked for uniqueness)
  - **Username**
  - **Password** (encrypted when stored)
- If the email is already taken, an error response is returned.
- The forum checks the credentials upon login to verify the password.

### Communication
- Only registered users can create posts and comments.
- Posts can be categorized, and categories can be chosen during post creation.
- Posts and comments are visible to both registered and non-registered users.

### Likes and Dislikes
- Registered users can like or dislike posts and comments.
- The number of likes and dislikes is visible to all users.

### Filter Mechanism
Users can filter posts by:

- **Categories** (subforums)
- **Created posts** (only for registered users)
- **Liked posts** (only for registered users)

### Docker
- The application is containerized using **Docker** for easier deployment and compatibility across environments.

## Setup Instructions

### Prerequisites
1. **Docker** installed on your machine.
2. **Go** (Golang) for building and running the server.

### Running the Project

1. **Clone the repository**:
   ```bash
   git clone https://github.com/Psyking0/FORUM
   cd forum-project
2. **Build the Docker image**:
   ```bash
    docker build -t forum-image .

3. **Run the Docker container:**:
   ```bash
    docker run -d -p 8080:8080 --name forum-container forum-image

