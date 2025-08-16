package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Peminjam struct {
	ID         int    `json:"id"`
	User       string `json:"user"`
	NoTelepon  string `json:"no_telepon"`
	Buku       string `json:"buku"`
	Tanggal    string `json:"tanggal"`
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:Asgar@123@tcp(127.0.0.1:3306)/perpustakaan")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Database tidak dapat diakses:", err)
	}
}

func getPeminjam(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM peminjam")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var list []Peminjam
	for rows.Next() {
		var p Peminjam
		var teleponBytes, tanggalBytes []byte
		err := rows.Scan(&p.ID, &p.User, &teleponBytes, &p.Buku, &tanggalBytes)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		p.NoTelepon = string(teleponBytes)
		p.Tanggal = string(tanggalBytes)
		list = append(list, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func createPeminjam(w http.ResponseWriter, r *http.Request) {
	var p Peminjam
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	stmt, err := db.Prepare("INSERT INTO peminjam (user, no_telepon, buku, tanggal) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(p.User, p.NoTelepon, p.Buku, p.Tanggal)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	id, _ := result.LastInsertId()
	p.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func updatePeminjam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID tidak valid", 400)
		return
	}

	var p Peminjam
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	stmt, err := db.Prepare("UPDATE peminjam SET user = ?, no_telepon = ?, buku = ?, tanggal = ? WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(p.User, p.NoTelepon, p.Buku, p.Tanggal, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Data peminjam berhasil diupdate")
}

func deletePeminjam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID tidak valid", 400)
		return
	}

	stmt, err := db.Prepare("DELETE FROM peminjam WHERE id = ?")
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
	fmt.Fprintln(w, "Data peminjam berhasil dihapus")
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}


func main() {
	initDB()
	router := mux.NewRouter()

	router.HandleFunc("/api/peminjam", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getPeminjam(w, r)
		case "POST":
			createPeminjam(w, r)
		default:
			http.Error(w, "Metode tidak didukung", http.StatusMethodNotAllowed)
		}
	})

	router.HandleFunc("/api/peminjam/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			updatePeminjam(w, r)
		case "DELETE":
			deletePeminjam(w, r)
		default:
			http.Error(w, "Metode tidak didukung", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server berjalan di http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", enableCORS(router)))

}
