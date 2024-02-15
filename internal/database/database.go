package internal

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

var Db *sqlx.DB

func NewPostgres() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	Db, err = sqlx.Connect("postgres", psqlInfo)
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
		user_id INT ,
		FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			comment VARCHAR(500) ,
			post_id INT ,
			user_id INT ,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id SERIAL PRIMARY KEY,
			post_id INT ,
			tag_name VARCHAR(50) ,
			FOREIGN KEY (post_id) REFERENCES posts(id)
		)`,
	}

	_, err := Db.Exec(queryDrop)
	if err != nil {
		log.Fatal("Error while dropping tables")
	}

	for _, q := range queriesCreate {
		_, err := Db.Exec(q)
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("MIRGATION FINISHED")

}
