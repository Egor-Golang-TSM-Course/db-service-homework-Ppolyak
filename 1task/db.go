package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

// User представляет собой структуру для таблицы users
type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Token    string
}

// SessionStorage представляет собой структуру для таблицы session_storage
type SessionStorage struct {
	ID     int    `db:"id"`
	UserID int    `db:"user_id"`
	Token  string `db:"token"`
}

// Post представляет собой структуру для таблицы posts
type Post struct {
	ID      int    `db:"id"`
	Title   string `db:"title"`
	Content string `db:"content"`
	UserID  int    `db:"user_id"`
}

// Comment представляет собой структуру для таблицы comments
type Comment struct {
	ID     int `db:"id"`
	PostID int `db:"post_id"`
	UserID int `db:"user_id"`
}

// Tag представляет собой структуру для таблицы tags
type Tag struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

/*func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if isValidToken(token) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}*/

/*func isValidToken(token string) bool {
	return token != ""
}*/

var db *sqlx.DB

func newPostgres() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db, err = sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	migrateDb()
}

func migrateDb() {
	queryDrop := `DROP TABLE IF EXISTS users, session_storage, comments, posts, tags`

	queriesCreate := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			password VARCHAR(50) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS session_storage (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			token VARCHAR(1000) NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
		id SERIAL PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		content TEXT NOT NULL,
		user_id INT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			post_id INT ,
			user_id INT ,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) 
		)`,
	}

	_, err := db.Exec(queryDrop)
	if err != nil {
		log.Fatal("Error while dropping tables")
	}

	for _, q := range queriesCreate {
		db.Exec(q)
	}

	log.Println("MIRGATION FINISHED")

}
