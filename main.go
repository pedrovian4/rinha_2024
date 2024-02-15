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
	return db, nil
}

func main() {
	db, err := initDB()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}(db)
	if err != nil {
		return
	}
	var h Handler
	h.conn = db
	http.HandleFunc("POST /clientes/{id}/transacoes", h.Transaction)
	http.HandleFunc("GET /clientes/{id}/extrato", h.BankStmt)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
