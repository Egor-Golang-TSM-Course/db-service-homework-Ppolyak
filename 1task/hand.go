package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"io"
	"net/http"
)

func registerUser(w http.ResponseWriter, r *http.Request) {
	var user User

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &user); err != nil {
		http.Error(w, "Error while decoding JSON", http.StatusBadRequest)
		return
	}

	if user.Username == "" {
		http.Error(w, "Missing username in JSON", http.StatusBadRequest)
		return
	}

	if user.Password == "" {
		http.Error(w, "Missing username in JSON", http.StatusBadRequest)
		return
	}

	if len(user.Username) < 5 {
		http.Error(w, "Username must be at least 5 characters long", http.StatusBadRequest)
		return
	}

	db.Create(&user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Error while encoding JSON", http.StatusInternalServerError)
	}

	/*generateToken(user, "secret")*/
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

	if trimmedUsername := user.Username; trimmedUsername == "" || user.Password == "" {
		http.Error(w, "Username and Password cannot be empty", http.StatusBadRequest)
		return
	}

	var logUser User
	if err := db.Where("username = ?", user.Username).First(&logUser).Error; err != nil {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	if logUser.Password != user.Password {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(logUser, "your_secret_key")
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	logUser.Token = token

	session := SessionStorage{
		UserID: logUser.ID,
		Token:  token,
	}
	db.Create(&session)

	db.Save(&logUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(logUser); err != nil {
		http.Error(w, "Error while encoding JSON", http.StatusInternalServerError)
	}
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
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	post.UserID = user.ID

	if post.Title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	db.Create(&post)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

}

func getAllPosts(w http.ResponseWriter, r *http.Request) {
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

}
