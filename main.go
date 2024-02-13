package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"os"
	"time"
)

func initDB() *pgxpool.Pool {
	//dbUrl := "postgres://rinha:rinha@localhost:5432/rinha"
	dbUrl := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(dbUrl)
	config.MaxConns = 100

	if err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Second)
	conn, _ := pgxpool.ConnectConfig(context.Background(), config)
	err = conn.Ping(context.Background())
	for err != nil {
		conn, _ = pgxpool.ConnectConfig(context.Background(), config)
		err = conn.Ping(context.Background())
	}
	return conn
}
func main() {
	conn := initDB()
	var h Handler
	h.conn = conn
	defer conn.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /clientes/{id}/transacoes", h.Transaction)
	mux.HandleFunc("GET /clientes/{id}/extrato", h.BankStmt)
	http.ListenAndServe(":8080", mux)
}
