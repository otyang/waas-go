package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Wallet struct {
	WalletID         string  `json:"wallet_id"`
	CustomerID       string  `json:"customer_id"`
	Currency         string  `json:"currency"`
	Amount           float64 `json:"amount"`
	IsActive         bool    `json:"is_active"`
	AvailableBalance float64 `json:"available_balance"`
	PendingBalance   float64 `json:"pending_balance"`
}

type CreditRequest struct {
	WalletID string  `json:"wallet_id"`
	Amount   float64 `json:"amount"`
}

type DebitRequest struct {
	WalletID string  `json:"wallet_id"`
	Amount   float64 `json:"amount"`
}

type TransferRequest struct {
	SourceWalletID      string  `json:"source_wallet_id"`
	DestinationWalletID string  `json:"destination_wallet_id"`
	Amount              float64 `json:"amount"`
}

type SwapRequest struct {
	FirstWalletID  string  `json:"first_wallet_id"`
	SecondWalletID string  `json:"second_wallet_id"`
	Amount         float64 `json:"amount"`
}

type LockRequest struct {
	WalletID string `json:"wallet_id"`
	Lock     bool   `json:"lock"`
}

var db *sql.DB

func initDB() {
	var err error
	connStr := "user=postgres dbname=walletdb password=postgres sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to database")
}

func main() {
	initDB()
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/wallets", createWallet).Methods("POST")
	router.HandleFunc("/wallets/{id}", getWallet).Methods("GET")
	router.HandleFunc("/wallets/credit", creditWallet).Methods("POST")
	router.HandleFunc("/wallets/debit", debitWallet).Methods("POST")
	router.HandleFunc("/wallets/transfer", transferFunds).Methods("POST")
	router.HandleFunc("/wallets/swap", swapBalances).Methods("POST")
	router.HandleFunc("/wallets/lock", lockWallet).Methods("POST")

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func createWallet(w http.ResponseWriter, r *http.Request) {
	var wallet Wallet
	err := json.NewDecoder(r.Body).Decode(&wallet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wallet.WalletID = generateUUID()
	wallet.IsActive = true
	wallet.AvailableBalance = wallet.Amount
	wallet.PendingBalance = 0

	stmt := `INSERT INTO wallets (wallet_id, customer_id, currency, amount, is_active, available_balance, pending_balance) 
	         VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = db.Exec(stmt, wallet.WalletID, wallet.CustomerID, wallet.Currency,
		wallet.Amount, wallet.IsActive, wallet.AvailableBalance, wallet.PendingBalance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

func getWallet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	walletID := params["id"]

	var wallet Wallet
	row := db.QueryRow("SELECT wallet_id, customer_id, currency, amount, is_active, available_balance, pending_balance FROM wallets WHERE wallet_id = $1", walletID)

	err := row.Scan(&wallet.WalletID, &wallet.CustomerID, &wallet.Currency, &wallet.Amount,
		&wallet.IsActive, &wallet.AvailableBalance, &wallet.PendingBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Wallet not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

func creditWallet(w http.ResponseWriter, r *http.Request) {
	var req CreditRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if wallet exists and is active
	var isActive bool
	err = tx.QueryRow("SELECT is_active FROM wallets WHERE wallet_id = $1 FOR UPDATE", req.WalletID).Scan(&isActive)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			http.Error(w, "Wallet not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isActive {
		tx.Rollback()
		http.Error(w, "Wallet is locked", http.StatusForbidden)
		return
	}

	// Update wallet balance
	_, err = tx.Exec("UPDATE wallets SET amount = amount + $1, available_balance = available_balance + $1 WHERE wallet_id = $2",
		req.Amount, req.WalletID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully credited %.2f to wallet %s", req.Amount, req.WalletID)
}

func debitWallet(w http.ResponseWriter, r *http.Request) {
	var req DebitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if wallet exists, is active, and has sufficient balance
	var isActive bool
	var availableBalance float64
	err = tx.QueryRow("SELECT is_active, available_balance FROM wallets WHERE wallet_id = $1 FOR UPDATE", req.WalletID).Scan(&isActive, &availableBalance)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			http.Error(w, "Wallet not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isActive {
		tx.Rollback()
		http.Error(w, "Wallet is locked", http.StatusForbidden)
		return
	}

	if availableBalance < req.Amount {
		tx.Rollback()
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	// Update wallet balance
	_, err = tx.Exec("UPDATE wallets SET amount = amount - $1, available_balance = available_balance - $1 WHERE wallet_id = $2",
		req.Amount, req.WalletID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully debited %.2f from wallet %s", req.Amount, req.WalletID)
}

func transferFunds(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check source wallet
	var sourceIsActive bool
	var sourceAvailableBalance float64
	var sourceCurrency string
	err = tx.QueryRow("SELECT is_active, available_balance, currency FROM wallets WHERE wallet_id = $1 FOR UPDATE", req.SourceWalletID).
		Scan(&sourceIsActive, &sourceAvailableBalance, &sourceCurrency)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			http.Error(w, "Source wallet not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !sourceIsActive {
		tx.Rollback()
		http.Error(w, "Source wallet is locked", http.StatusForbidden)
		return
	}

	if sourceAvailableBalance < req.Amount {
		tx.Rollback()
		http.Error(w, "Insufficient funds in source wallet", http.StatusBadRequest)
		return
	}

	// Check destination wallet
	var destIsActive bool
	var destCurrency string
	err = tx.QueryRow("SELECT is_active, currency FROM wallets WHERE wallet_id = $1 FOR UPDATE", req.DestinationWalletID).
		Scan(&destIsActive, &destCurrency)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			http.Error(w, "Destination wallet not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !destIsActive {
		tx.Rollback()
		http.Error(w, "Destination wallet is locked", http.StatusForbidden)
		return
	}

	if sourceCurrency != destCurrency {
		tx.Rollback()
		http.Error(w, "Cannot transfer between different currencies", http.StatusBadRequest)
		return
	}

	// Perform transfer
	_, err = tx.Exec("UPDATE wallets SET amount = amount - $1, available_balance = available_balance - $1 WHERE wallet_id = $2",
		req.Amount, req.SourceWalletID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE wallets SET amount = amount + $1, available_balance = available_balance + $1 WHERE wallet_id = $2",
		req.Amount, req.DestinationWalletID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully transferred %.2f from wallet %s to wallet %s",
		req.Amount, req.SourceWalletID, req.DestinationWalletID)
}

func swapBalances(w http.ResponseWriter, r *http.Request) {
	var req SwapRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check first wallet
	var firstIsActive bool
	var firstAvailableBalance float64
	err = tx.QueryRow("SELECT is_active, available_balance FROM wallets WHERE wallet_id = $1 FOR UPDATE", req.FirstWalletID).
		Scan(&firstIsActive, &firstAvailableBalance)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			http.Error(w, "First wallet not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !firstIsActive {
		tx.Rollback()
		http.Error(w, "First wallet is locked", http.StatusForbidden)
		return
	}

	if firstAvailableBalance < req.Amount {
		tx.Rollback()
		http.Error(w, "Insufficient funds in first wallet", http.StatusBadRequest)
		return
	}

	// Check second wallet
	var secondIsActive bool
	var secondAvailableBalance float64
	err = tx.QueryRow("SELECT is_active, available_balance FROM wallets WHERE wallet_id = $1 FOR UPDATE", req.SecondWalletID).
		Scan(&secondIsActive, &secondAvailableBalance)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			http.Error(w, "Second wallet not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !secondIsActive {
		tx.Rollback()
		http.Error(w, "Second wallet is locked", http.StatusForbidden)
		return
	}

	// Perform swap
	_, err = tx.Exec("UPDATE wallets SET amount = amount - $1, available_balance = available_balance - $1 WHERE wallet_id = $2",
		req.Amount, req.FirstWalletID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE wallets SET amount = amount + $1, available_balance = available_balance + $1 WHERE wallet_id = $2",
		req.Amount, req.SecondWalletID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully swapped %.2f between wallet %s and wallet %s",
		req.Amount, req.FirstWalletID, req.SecondWalletID)
}

func lockWallet(w http.ResponseWriter, r *http.Request) {
	var req LockRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if wallet exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM wallets WHERE wallet_id = $1)", req.WalletID).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	}

	// Update lock status
	_, err = db.Exec("UPDATE wallets SET is_active = $1 WHERE wallet_id = $2", !req.Lock, req.WalletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	action := "unlocked"
	if req.Lock {
		action = "locked"
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully %s wallet %s", action, req.WalletID)
}

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
