package internal

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

type PostgresDB struct {
	db *sqlx.DB
}

func NewPostgres() (*PostgresDB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	p := &PostgresDB{db: db}
	err = p.MigrateDb()
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return p, nil
}

func (p *PostgresDB) MigrateDb() error {
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

	_, err := p.db.Exec(queryDrop)
	if err != nil {
		return fmt.Errorf("error while dropping tables: %w", err)
	}

	for _, q := range queriesCreate {
		_, err := p.db.Exec(q)
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("Migration finished")

	return nil
}
