package main

import (
	_ "encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "io"
	"log"
	"net/http"
	"time"
)

func main() {
	newPostgres()

	r := mux.NewRouter()
	r.HandleFunc("/users/register", registerUser).Methods("POST")
	r.HandleFunc("/users/login", loginUser).Methods("POST")
	r.HandleFunc("/posts", createPost).Methods("POST")
	//r.HandleFunc("/posts", getAllPosts).Methods("GET")
	//r.HandleFunc("/posts/{id}", getPostById).Methods("GET")
	//r.HandleFunc("/posts/{id}", updatePost).Methods("PUT")

	port := ":6262"
	fmt.Printf("Сервер запущен на порту %s...\n", port)

	http.ListenAndServe(port, r)
}

func generateToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID, // Предполагается, что ID пользователя является числом
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("your_secret_key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func authenticateUser(r *http.Request) (*User, error) {
	token := r.Header.Get("Authorization")

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("your_secret_key"), nil
	})
	if err != nil {
		return nil, err
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return nil, errors.New("invalid token")
	}

	userID := int(sub)

	var session SessionStorage
	log.Println(userID)
	err = db.Get(&session, "SELECT * FROM session_storage WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	var user User
	err = db.Get(&user, "SELECT * FROM users WHERE id = $1", session.UserID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
