package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx"
	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"

	"github.com/vladyslav2/bitcoin2sql/pkg/address"
	"github.com/vladyslav2/bitcoin2sql/pkg/block"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// App holding routers and DB connection
type App struct {
	Router *mux.Router
	DB     *pgx.ConnPool
}

// Initialize application and open db connection
func (a *App) Initialize(pg *pgx.ConnPool) {

	a.Router = mux.NewRouter()
	a.DB = pg
	a.initializeRoutes()
}

// Run application on 8080 port
func (a *App) Run(addr string) {

	if addr == "" {
		addr = "8000"
	}

	log.Fatal(http.ListenAndServe(":8000", a.Router))
}

// initializeRoutes - creates routers, runs automatically in Initialize
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/", a.mainPage).Methods("GET")
	a.Router.HandleFunc("/{hash:[0-9a-zA-Z]+}", a.showBlock).Methods("GET")
	a.Router.HandleFunc("/address/{hash:[0-9a-zA-Z]+}", a.showAddress).Methods("GET")
}

func (a *App) mainPage(w http.ResponseWriter, r *http.Request) {
	storage := address.NewStorage(a.DB)
	last10, err := storage.Last10("")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	rich10, err := storage.MostRich()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	res := map[string]interface{}{
		"last10": last10,
		"rich10": rich10,
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (a *App) showBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	storage := block.NewStorage(a.DB)
	b, err := storage.GetByHash(vars["hash"])
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Block not found")
			return
		}
		respondWithError(w, http.StatusBadRequest, "Cannot retrieve block")
		return
	}

	if _, err = b.GetTransactions(); err != nil {
		log.Printf("error in block gettransactions, %v", err)
		respondWithError(w, http.StatusNotFound, "block not found")
		return
	}

	if err := b.GetPrice(); err != nil {
		log.Printf("error in block get price, %v", err)
		respondWithError(w, http.StatusServiceUnavailable, "Prices for block not found")
		return
	}
	respondWithJSON(w, http.StatusOK, b)
}

func (a *App) showAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	storage := address.NewStorage(a.DB)
	addr := address.New(&storage)
	if err := addr.GetByHash(vars["hash"]); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Address not found")
		} else {
			log.Printf("error during show address, %v", err)
			respondWithError(w, http.StatusBadRequest, "Cannot retrieve address")
		}
		return
	}

	if err := addr.GetTransactions(); err != nil {
		log.Printf("app: error in address gettransactions, %v", err)
		respondWithError(w, http.StatusServiceUnavailable, "Cannot get transactions for address")
		return
	}
	if err := transaction.GetPricePerTransaction(
		transaction.NewStorage(a.DB),
		addr.Transactions,
	); err != nil {
		log.Printf("app: error in address getprices, %v", err)
		respondWithError(w, http.StatusServiceUnavailable, "Prices for transactions not found")
		return
	}

	respondWithJSON(w, http.StatusOK, addr)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)

	if err != nil {
		log.Fatalf("Cannot convert data to json, %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	respondWithBytes(w, code, response)
}

func respondWithBytes(w http.ResponseWriter, code int, response []byte) {
	w.WriteHeader(code)
	w.Write(response)
}
