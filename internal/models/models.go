package models

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
	ID      int    `db:"id"`
	Comment string `db:"comment"`
	PostID  int    `db:"post_id"`
	UserID  int    `db:"user_id"`
}

// Tag представляет собой структуру для таблицы tags
type Tag struct {
	ID      int    `db:"id"`
	PostID  int    `db:"post_id"`
	TagName string `db:"tag_name"`
}
