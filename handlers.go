package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	conn *sql.DB
}

// BEGIN TRANSACTION

func validateTransaction(body io.Reader) (TransactionRequest, error) {

	var tRequest TransactionRequest
	err := json.NewDecoder(body).Decode(&tRequest)
	if err != nil {
		return tRequest, err
	}
	if "c" != tRequest.Type && "d" != tRequest.Type {
		return tRequest, fmt.Errorf("type not available")
	}

	if len(tRequest.Description) > 10 {
		return tRequest, fmt.Errorf("invalid description")
	}
	if len(tRequest.Description) < 1 {
		return tRequest, fmt.Errorf("invalid description")
	}

	return tRequest, nil
}

func (h *Handler) Transaction(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	id, err := strconv.ParseUint(request.PathValue("id"), 10, 3)
	if err != nil {
		if strings.Contains(err.Error(), "value out of range") {
			http.Error(writer, "Number is too large", http.StatusNotFound)
		} else {
			http.Error(writer, "Invalid ID", http.StatusUnprocessableEntity)
		}
		return
	}
	tRequest, err := validateTransaction(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	var c Clients

	err = h.conn.QueryRow("SELECT id, name, \"limit\" , balance FROM clients where id = $1", id).Scan(&c.Id, &c.Name, &c.Limit, &c.Balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		http.Error(writer, fmt.Sprintf("Internal Server Error %s", err.Error()), http.StatusInternalServerError)
		return
	}
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
		c.Balance += tRequest.Value
	}

	_, err = h.conn.Exec("UPDATE clients set balance = $1 WHERE id = $2", c.Balance, c.Id)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}
	_, err = h.conn.Exec("INSERT INTO transactions (value, type, description, created_at, client_id)  VALUES ($1,$2,$3,$4, $5)", tRequest.Value, tRequest.Type, tRequest.Description, time.Now(), id)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(writer).Encode(map[string]int{
		"limite": c.Limit,
		"saldo":  c.Balance,
	})
}

// END TRANSACTION

// BankStmt START
func (h *Handler) BankStmt(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	id, err := strconv.ParseUint(request.PathValue("id"), 10, 3)
	if err != nil {
		if strings.Contains(err.Error(), "value out of range") {
			http.Error(writer, "Number is too large", http.StatusNotFound)
		} else {
			http.Error(writer, "Invalid ID", http.StatusUnprocessableEntity)
		}
		return
	}
	var balance, limit int
	err = h.conn.QueryRow("SELECT balance, \"limit\" FROM clients WHERE id = $1", id).Scan(&balance, &limit)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	var stmt BankStmt

	stmt.Balance = Balance{
		Total:       balance,
		BalanceDate: time.Now(),
		Limit:       limit,
	}
	rows, err := h.conn.Query("SELECT value, type, description, created_at FROM transactions WHERE client_id = $1 ORDER BY created_at DESC LIMIT 10", id)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writer.WriteHeader(404)
			return
		} else {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		counter := 0
		for rows.Next() {
			var t Transaction
			err := rows.Scan(&t.Value, &t.Type, &t.Description, &t.CreatedAt)
			if err != nil {
				return
			}
			stmt.LastTransactions[counter] = t
			counter++
		}
	}

	err = json.NewEncoder(writer).Encode(stmt)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

// END BANK STMT
