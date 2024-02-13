package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	conn *pgxpool.Pool
}

// BEGIN TRANSACTION

func validateTransaction(request http.Request) (*TransactionRequest, error) {

	var tRequest TransactionRequest
	err := json.NewDecoder(request.Body).Decode(&tRequest)
	if err != nil {
		return nil, err
	}
	if "c" != tRequest.Type && "d" != tRequest.Type {
		return nil, fmt.Errorf("type not available")
	}

	if len(tRequest.Description) > 10 {
		return nil, fmt.Errorf("invalid description")
	}
	if len(tRequest.Description) < 1 {
		return nil, fmt.Errorf("invalid description")
	}

	return &tRequest, nil
}

func (h *Handler) Transaction(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(request.PathValue("id"))
	if err != nil {
		fmt.Println("emcima")

		http.Error(writer, "Invalid ID", http.StatusUnprocessableEntity)
		return
	}

	tRequest, err := validateTransaction(*request)
	if err != nil {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Println("aqui")
		return
	}
	var c Clients
	err = h.conn.QueryRow(context.Background(), "SELECT id, name, \"limit\" , balance FROM clients where id = $1", id).Scan(&c.Id, &c.Name, &c.Limit, &c.Balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("Not found")
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		http.Error(writer, fmt.Sprintf("Internal Server Error %s", err.Error()), http.StatusInternalServerError)
		return
	}
	tx, err := h.conn.Begin(context.Background())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if tRequest.Type == "d" {
		if c.Balance-tRequest.Value < -c.Limit {
			http.Error(writer, "Transaction would result in an inconsistent balance", http.StatusUnprocessableEntity)
			return
		}
		c.Balance -= tRequest.Value

	} else if tRequest.Type == "c" {
		c.Balance -= tRequest.Value
	}

	_, err = h.conn.Exec(context.Background(), "UPDATE clients set balance = $1 WHERE id = $2", c.Balance, c.Id)
	if err != nil {
		_ = tx.Rollback(context.Background())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = h.conn.Exec(context.Background(), "INSERT INTO transactions (value, type, description, created_at, client_id)  VALUES ($1,$2,$3,$4, $5)", tRequest.Value, tRequest.Type, tRequest.Description, time.Now(), id)
	if err != nil {
		_ = tx.Rollback(context.Background())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tx.Commit(context.Background())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(writer).Encode(map[string]int{
		"limite": c.Limit,
		"saldo":  c.Balance,
	})
	writer.WriteHeader(http.StatusOK)
}

// END TRANSACTION

// START BANK STMT
func (h *Handler) BankStmt(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(request.PathValue("id"))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var balance, limit int
	err = h.conn.QueryRow(context.Background(), "SELECT balance, \"limit\" FROM clients WHERE id = $1", id).Scan(&balance, &limit)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	var stmt BankStmt

	stmt.Balance = map[string]any{
		"total":        balance,
		"data_extrato": time.Now(),
		"limite":       limit,
	}
	rows, err := h.conn.Query(context.Background(), "SELECT value, type, description, created_at FROM transactions WHERE client_id = $1 ORDER BY created_at DESC LIMIT 10", id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			stmt.LastTransactions = make([]Transaction, 0)
			return
		} else {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		for rows.Next() {
			var t Transaction
			err := rows.Scan(&t.Value, &t.Type, &t.Description, &t.CreatedAt)
			if err != nil {
				return
			}
			stmt.LastTransactions = append(stmt.LastTransactions, t)
		}
	}

	err = json.NewEncoder(writer).Encode(stmt)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	writer.WriteHeader(200)

}

// END BANK STMT
