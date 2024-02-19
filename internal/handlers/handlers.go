package handlers

import (
	. "awesomeProject15/db-service-homework-Ppolyak/internal/database"
	. "awesomeProject15/db-service-homework-Ppolyak/internal/models"
	"awesomeProject15/db-service-homework-Ppolyak/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	_ "github.com/jmoiron/sqlx"
	"io"
	_ "io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
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

	_, err = Db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	if err != nil {
		http.Error(w, "Error while inserting into users table", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
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

	err = Db.Get(&user, "SELECT id, username FROM users WHERE username = $1 AND password = $2", user.Username, user.Password)
	if err != nil {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	user.Token = token

	_, err = Db.Exec("INSERT INTO session_storage (user_id, token) VALUES ($1, $2)", user.ID, user.Token)
	if err != nil {
		log.Println("Error updating session storage:", err)
		http.Error(w, "Error updating session storage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post Post
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	user := r.Context().Value("user").(*User) // Получаем пользователя из контекста запроса

	if err := json.Unmarshal(body, &post); err != nil {
		http.Error(w, "Error while decoding JSON", http.StatusBadRequest)
		return
	}

	post.UserID = user.ID
	log.Println("Authenticated user:", user.ID)

	if post.Title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = Db.Exec("INSERT INTO posts (title, content, user_id) VALUES ($1, $2, $3)", post.Title, post.Content, post.UserID)
	if err != nil {
		log.Println("Error creating post:", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

func GetAllPosts(w http.ResponseWriter, r *http.Request) {
	var post []Post
	user := r.Context().Value("user").(*User)

	err := Db.Select(&post, "SELECT * FROM posts WHERE user_id = $1", user.ID)
	log.Println(user.ID)
	if err != nil {
		log.Println("Error while getting all posts:", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

}

func GetPostById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId, ok := vars["id"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(postId)
	if err != nil {
		log.Println(err)
	}

	var post Post

	if err := Db.Get(&post, "SELECT * FROM posts WHERE id = $1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "Error encoding json", http.StatusInternalServerError)
	}
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	vars := mux.Vars(r)
	postId, ok := vars["id"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}

	var post Post

	err := Db.Get(&post, "SELECT * FROM posts WHERE ID = $1", postId)

	if post.UserID != user.ID {
		http.Error(w, "It's not your post", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request body", http.StatusInternalServerError)
	}

	defer r.Body.Close()

	if err := json.Unmarshal(body, &post); err != nil {
		http.Error(w, "Error while unmarshal", http.StatusBadRequest)
		return
	}

	if post.Title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = Db.Exec("UPDATE posts SET title = $1, content = $2 WHERE ID = $3;", post.Title, post.Content, postId)
	if err != nil {
		http.Error(w, "Error while updating post", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	vars := mux.Vars(r)
	postId, ok := vars["id"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}

	var post Post

	err := Db.Get(&post, "SELECT * FROM posts WHERE ID = $1", postId)

	if post.UserID != user.ID {
		http.Error(w, "It's not your post", http.StatusUnauthorized)
		return
	}

	_, err = Db.Exec("DELETE FROM posts WHERE ID = $1", postId)
	if err != nil {
		http.Error(w, "Error while deleting post", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deleted"))
}

func AddCommentToPost(w http.ResponseWriter, r *http.Request) {
	var comment Comment
	vars := mux.Vars(r)
	postId, ok := vars["id"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	user := r.Context().Value("user").(*User) // Получаем пользователя из контекста запроса

	if err := json.Unmarshal(body, &comment); err != nil {
		http.Error(w, "Error while decoding JSON", http.StatusBadRequest)
		return
	}

	if comment.Comment == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = Db.Exec("INSERT INTO comments (comment, post_id, user_id) VALUES ($1, $2, $3)", comment.Comment, postId, user.ID)
	if err != nil {
		log.Println("Error creating comment:", err)
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(comment); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId, ok := vars["id"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}
	var comment []Comment

	err := Db.Select(&comment, "SELECT * FROM comments WHERE post_id = $1", postId)
	if err != nil {
		http.Error(w, "Error while getting post", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(comment); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

func AddTag(w http.ResponseWriter, r *http.Request) {
	var tag Tag
	var post Post

	vars := mux.Vars(r)
	postId, ok := vars["postId"]
	if !ok {
		http.Error(w, "Post id is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	user := r.Context().Value("user").(*User)

	err = Db.Get(&post, "SELECT * FROM posts WHERE ID = $1", postId)

	if err := json.Unmarshal(body, &tag); err != nil {
		http.Error(w, "Error while decoding JSON", http.StatusBadRequest)
		return
	}

	if post.UserID != user.ID {
		http.Error(w, "It's not your post", http.StatusUnauthorized)
		return
	}

	if tag.TagName == "" {
		http.Error(w, "Tag name can't be empty", http.StatusBadRequest)
		return
	}

	_, err = Db.Exec("INSERT INTO tags (post_id, tag_name) VALUES ($1,$2)", postId, tag.TagName)
	if err != nil {
		http.Error(w, "Error while adding tag", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(tag); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

}

func GetTags(w http.ResponseWriter, r *http.Request) {
	var tag []Tag
	pageS := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageS)
	if err != nil || page < 1 {
		page = 1
	}

	pageQS := r.URL.Query().Get("pageQ")
	pageQ, err := strconv.Atoi(pageQS)
	if err != nil || pageQ < 1 {
		pageQ = 10
	}

	offset := (page - 1) * pageQ

	err = Db.Select(&tag, "SELECT * FROM tags LIMIT $1 OFFSET $2", pageQ, offset)
	if err != nil {
		http.Error(w, "Error while getting post", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tag); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

func SearchForKeyWord(w http.ResponseWriter, r *http.Request) {
	var posts []Post
	keyword := r.URL.Query().Get("keyword")

	query := "SELECT * FROM posts WHERE title LIKE '%' || $1 || '%' OR content LIKE '%' || $1 || '%'"
	err := Db.Select(&posts, query, keyword)
	if err != nil {
		http.Error(w, "Error while searching posts", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}
