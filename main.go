package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var jwtKey = []byte("kunci_rahasia")

type User struct {
	ID       int    `json:"id"`
	Nama     string `json:"nama"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:Asgar@123@tcp(127.0.0.1:3306)/latihan_go")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Database tidak dapat diakses:", err)
	}
}

// ===== LOGIN =====
func login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow("SELECT id, nama, username, password, email FROM user WHERE username = ?", creds.Username).
		Scan(&user.ID, &user.Nama, &user.Username, &user.Password, &user.Email)
	if err != nil {
		http.Error(w, "User tidak ditemukan", http.StatusUnauthorized)
		return
	}

	if creds.Password != user.Password {
		http.Error(w, "Password salah", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// ===== MIDDLEWARE =====
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Token tidak ada", http.StatusUnauthorized)
			return
		}

		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Token tidak valid", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userClaims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ===== HANDLER =====
func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM user")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Nama, &u.Username, &u.Password, &u.Email)
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func getMyProfile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("userClaims").(jwt.MapClaims)
	username := claims["username"].(string)

	var u User
	err := db.QueryRow("SELECT id, nama, username, password, email FROM user WHERE username = ?", username).
		Scan(&u.ID, &u.Nama, &u.Username, &u.Password, &u.Email)
	if err != nil {
		http.Error(w, "User tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	stmt, err := db.Prepare("INSERT INTO user (nama, username, password, email) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(u.Nama, u.Username, u.Password, u.Email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	id, _ := result.LastInsertId()
	u.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID tidak valid", 400)
		return
	}

	stmt, err := db.Prepare("DELETE FROM user WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "User berhasil dihapus")
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID tidak valid", 400)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	stmt, err := db.Prepare("UPDATE user SET nama = ?, username = ?, password = ?, email = ? WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Nama, u.Username, u.Password, u.Email, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "User berhasil diupdate")
}

// ===== MAIN =====
func main() {
	initDB()
	router := mux.NewRouter()

	router.HandleFunc("/login", login).Methods("POST")

	// Endpoint profil user yang login
	router.Handle("/api/me", authMiddleware(http.HandlerFunc(getMyProfile)))

	 handler := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Authorization", "Content-Type"},
    }).Handler(router)

	 fmt.Println("Server berjalan di http://localhost:8000")
    log.Fatal(http.ListenAndServe(":8000", handler))

	// Endpoint semua user (hanya contoh, bisa dipakai untuk admin)
	router.Handle("/api/user", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getUsers(w, r)
		case "POST":
			createUser(w, r)
		default:
			http.Error(w, "Metode tidak didukung", http.StatusMethodNotAllowed)
		}
	})))

	router.Handle("/api/user/{id}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			updateUser(w, r)
		case "DELETE":
			deleteUser(w, r)
		default:
			http.Error(w, "Metode tidak didukung", http.StatusMethodNotAllowed)
		}
	})))

	fmt.Println("Server berjalan di http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
