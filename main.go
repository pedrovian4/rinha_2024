package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"time"
)

func initDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	//dbURL := "postgres://rinha:rinha@localhost:5432/rinha?sslmode=disable"
	time.Sleep(5 * time.Second)
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to the database.")
	return db, nil
}

func main() {
	db, err := initDB()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	if err != nil {
		panic(err)
	}
	var h Handler
	h.conn = db
	mux := http.NewServeMux()
	mux.HandleFunc("POST /clientes/{id}/transacoes", h.Transaction)
	mux.HandleFunc("GET /clientes/{id}/extrato", h.BankStmt)
	http.ListenAndServe(":8080", mux)
}
