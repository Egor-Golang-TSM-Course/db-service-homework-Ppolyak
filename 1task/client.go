package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

type User struct {
	gorm.Model
	ID       int
	Username string         `json:"username"`
	Password string         `json:"password"`
	Token    string         `gorm:"not null"`
	Session  SessionStorage `gorm:"foreignKey:UserID" json:"-"`
	Posts    []Post         `gorm:"foreignKey:UserID"` // Отношение к постам пользователя
}

type Post struct {
	gorm.Model
	ID      int
	Title   string `json:"title" gorm:"not null"`
	Content string `json:"content" gorm:"not null"`
	UserID  int    `json:"-"`
	User    User   `gorm:"foreignKey:UserID" json:"-"`
}

type Comment struct {
	gorm.Model
	PostID uint
	UserID uint
	/*Text   string
	User   User*/
}

type SessionStorage struct {
	gorm.Model
	UserID int    `gorm:"uniqueIndex;not null"`
	Token  string `gorm:"not null"`
}

type Tag struct {
	gorm.Model
	Name string
	/*Post []Post*/
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

var db *gorm.DB

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db, err = gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	/*db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	*/
	db.Debug().Migrator().DropTable(&User{}, &SessionStorage{}, &Comment{}, &Post{}, &SessionStorage{})

	db.AutoMigrate(&Post{}, &User{}, &Comment{}, &Tag{}, &SessionStorage{})

	r := mux.NewRouter()
	r.HandleFunc("/users/register", registerUser).Methods("POST")
	r.HandleFunc("/users/login", loginUser).Methods("POST")
	r.HandleFunc("/posts", createPost).Methods("POST")
	r.HandleFunc("/posts", getAllPosts).Methods("GET")
	r.HandleFunc("/posts/{id}", getPostById).Methods("GET")
	r.HandleFunc("/posts/{id}", updatePost).Methods("PUT")

	port := ":6262"
	fmt.Printf("Сервер запущен на порту %s...\n", port)

	http.ListenAndServe(port, r)
}

func generateToken(user User, secretKey string) (string, error) {

	claims := jwt.MapClaims{
		"sub": user,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Время жизни токена
		"iat": time.Now().Unix(),                     // Время создания токена
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "err", nil
	}

	return tokenString, err
}

func authenticateUser(r *http.Request) (*User, error) {
	token := r.Header.Get("Authorization")
	var session SessionStorage

	if err := db.Where("token = ?", token).First(&session).Error; err != nil {
		return nil, err
	}

	var user User

	if err := db.Where("id = ?", session.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
