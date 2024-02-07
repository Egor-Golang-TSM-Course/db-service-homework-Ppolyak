package main

import (
	"encoding/json"
	_ "github.com/jmoiron/sqlx"
	"io"
	_ "io/ioutil"
	"log"
	"net/http"
)

func registerUser(w http.ResponseWriter, r *http.Request) {
	var user User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &user); err != nil {
		http.Error(w, "Error while decoding JSON", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	db.MustExec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var user User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &user); err != nil {
		http.Error(w, "Error while decoding JSON", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	err = db.Get(&user, "SELECT id, username FROM users WHERE username = $1 AND password = $2", user.Username, user.Password)
	if err != nil {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	user.Token = token

	_, err = db.Exec("INSERT INTO session_storage (user_id, token) VALUES ($1, $2)", user.ID, user.Token)
	if err != nil {
		log.Println("Error updating session storage:", err)
		http.Error(w, "Error updating session storage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &post); err != nil {
		http.Error(w, "Error while decoding JSON", http.StatusBadRequest)
		return
	}

	user, err := authenticateUser(r)
	if err != nil {
		log.Println("Authentication failed:", err) // Логирование ошибки аутентификации
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	log.Println("Authenticated user:", user.ID) // Логирование идентификатора аутентифицированного пользователя

	post.UserID = user.ID

	if post.Title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO posts (title, content, user_id) VALUES ($1, $2, $3)", post.Title, post.Content, post.UserID)
	if err != nil {
		log.Println("Error creating post:", err) // Логирование ошибки создания поста
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		log.Println("Error encoding JSON:", err) // Логирование ошибки кодирования JSON
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

/*func getAllPosts(w http.ResponseWriter, r *http.Request) {
	user, err := authenticateUser(r)
	if err != nil {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var post []Post
	if err := db.Where("user_id = ?", user.ID).Find(&post).Error; err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

}

func getPostById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId, ok := vars["id"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}

	var post Post
	if err := db.First(&post, postId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "Error encoding json", http.StatusInternalServerError)
	}
}

func updatePost(w http.ResponseWriter, r *http.Request) {
	user, err := authenticateUser(r)
	if err != nil {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	postId, ok := vars["id"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}

	var post Post
	if err := db.First(&post, postId).Error; err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if post.UserID != user.ID {
		http.Error(w, "It's not your post", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request body", http.StatusInternalServerError)
	}

	defer r.Body.Close()

	var updatePostStruct struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(body, &updatePostStruct); err != nil {
		http.Error(w, "Error while unmarshal", http.StatusBadRequest)
		return
	}

	if updatePostStruct.Title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	post.Title = updatePostStruct.Title
	post.Content = updatePostStruct.Content

	if err := db.Save(&post).Error; err != nil {
		http.Error(w, "Error while updating post", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

}*/
