package main

import "time"

type Clients struct {
	Id      int    `json:"id"`
	Name    string `json:"nome"`
	Limit   int    `json:"limite"`
	Balance int    `json:"saldo"`
}

type Transaction struct {
	Id          int       `json:"id"`
	Value       int       `json:"valor"`
	Type        string    `json:"tipo"`
	Description string    `json:"descricao"`
	CreatedAt   time.Time `json:"realizado_em"`
}
type Balance struct {
	Total       int       `json:"total"`
	BalanceDate time.Time `json:"data_extrato"`
	Limit       int       `json:"limite"`
}
type BankStmt struct {
	Balance          Balance         `json:"saldo"`
	LastTransactions [10]Transaction `json:"ultimas_transacoes"`
}

type TransactionRequest struct {
	Value       int    `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
}
