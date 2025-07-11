package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"real-time-forum/internal/auth"
	"real-time-forum/internal/models"
	"real-time-forum/internal/services"
	"real-time-forum/internal/websocket"
	"strconv"
	"time"
)

type API struct {
	DB  *sql.DB
	Hub *websocket.Hub
}

func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userService := services.UserService{DB: a.DB}
	if err := userService.Register(&user); err != nil {
		log.Printf("Registration failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (a *API) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		EmailOrNickname string `json:"emailOrNickname"`
		Password        string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userService := services.UserService{DB: a.DB}
	token, err := userService.Login(credentials.EmailOrNickname, credentials.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (a *API) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	userService := services.UserService{DB: a.DB}
	if err := userService.Logout(cookie.Value); err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (a *API) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket.ServeWs(a.Hub, w, r, a.DB)
}

// Similar handlers for posts, comments, chat, etc.
func (a *API) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	user, err := a.getSessionUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var post struct {
		Title      string   `json:"title"`
		Content    string   `json:"content"`
		Categories []string `json:"categories"`
	}

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	postService := services.PostService{DB: a.DB}
	if err := postService.CreatePost(user.ID, post.Title, post.Content, post.Categories); err != nil {
		log.Printf("Create post failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Post created successfully"})
}

func (a *API) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	postService := services.PostService{DB: a.DB}
	posts, err := postService.GetPosts(limit, (page-1)*limit)
	if err != nil {
		log.Printf("Get posts failed: %v", err)
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(posts)
}

func (a *API) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	user, err := a.getSessionUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comment struct {
		PostID  int    `json:"postId"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	postService := services.PostService{DB: a.DB}
	if err := postService.CreateComment(comment.PostID, user.ID, comment.Content); err != nil {
		log.Printf("Create comment failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Comment created successfully"})
}

func (a *API) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := a.getSessionUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	otherUserID, _ := strconv.Atoi(r.URL.Query().Get("userId"))
	if otherUserID == 0 {
		http.Error(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	chatService := services.ChatService{DB: a.DB}
	messages, err := chatService.GetMessages(user.ID, otherUserID, 100, 0)
	if err != nil {
		log.Printf("Get messages failed: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(messages)
}

func (a *API) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	user, err := a.getSessionUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var message struct {
		ReceiverID int    `json:"receiverId"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	chatService := services.ChatService{DB: a.DB}
	if err := chatService.SaveMessage(user.ID, message.ReceiverID, message.Content); err != nil {
		log.Printf("Send message failed: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Broadcast via WebSocket
	msg := models.ChatMessage{
		Sender:    user.Nickname,
		Receiver:  strconv.Itoa(message.ReceiverID),
		Content:   message.Content,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	wsMessage := models.WebSocketMessage{
		Type:    "chat_message",
		Payload: msg,
	}
	messageBytes, _ := json.Marshal(wsMessage)
	a.Hub.broadcast <- messageBytes

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Message sent successfully"})
}

func (a *API) GetOnlineUsersHandler(w http.ResponseWriter, r *http.Request) {
	onlineUsers := a.Hub.GetOnlineUsers()
	json.NewEncoder(w).Encode(onlineUsers)
}

// Helper to get current user from session
func (a *API) getSessionUser(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}
	return auth.ValidateSession(a.DB, cookie.Value)
}
