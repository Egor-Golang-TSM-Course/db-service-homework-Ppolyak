package main

import (
	. "awesomeProject15/db-service-homework-Ppolyak/internal/database"
	. "awesomeProject15/db-service-homework-Ppolyak/internal/handlers"
	. "awesomeProject15/db-service-homework-Ppolyak/internal/middleware"
	_ "encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "io"
	"net/http"
)

func main() {
	NewPostgres()

	r := mux.NewRouter()
	r.HandleFunc("/users/register", RegisterUser).Methods("POST")
	r.HandleFunc("/users/login", LoginUser).Methods("POST")
	r.HandleFunc("/posts", AuthMiddleware(CreatePost)).Methods("POST")
	r.HandleFunc("/posts", AuthMiddleware(GetAllPosts)).Methods("GET")
	r.HandleFunc("/posts/{id}", AuthMiddleware(GetPostById)).Methods("GET")
	r.HandleFunc("/posts/{id}", AuthMiddleware(UpdatePost)).Methods("PUT")
	r.HandleFunc("/posts/{id}", AuthMiddleware(DeletePost)).Methods("DELETE")
	//POST /posts/{postId}/comments
	r.HandleFunc("/posts/{id}/comments", AuthMiddleware(AddCommentToPost)).Methods("POST")
	r.HandleFunc("/posts/{id}/comments", AuthMiddleware(GetComments)).Methods("GET")
	port := ":6262"
	fmt.Printf("Сервер запущен на порту %s...\n", port)

	http.ListenAndServe(port, r)
}
