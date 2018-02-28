package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// App holding routers and DB connection
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize application and open db connection
func (a *App) Initialize(host, user, password, dbname string) {

	if host == "" {
		log.Fatal("Empty host string, setup DB_HOST env")
		host = "localhost"
	}

	if user == "" {
		log.Fatal("Empty user string, setup DB_USER env")
		return
	}

	if dbname == "" {
		log.Fatal("Empty dbname string, setup DB_DBNAME env")
		return
	}

	connectionString :=
		fmt.Sprintf("host=%s user=%s password='%s' dbname=%s sslmode=disable", host, user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Cannot open postgresql connection: %v", err)
	}
	//defer a.DB.Close()

	a.Router = mux.NewRouter()
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
	blocks := Last10(a.DB)
	respondWithJSON(w, http.StatusOK, blocks)
}

func (a *App) showBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	b := NewBlock(a.DB)
	err := b.getByHash(vars["hash"])
	if err != nil {
		if err.Error() == "404" {
			respondWithError(w, http.StatusNotFound, "Block not found")
			return
		}
		log.Fatalf("error in block getbyhash, %v", err)
		respondWithError(w, http.StatusBadRequest, "Cannot retrieve block")
		return
	}

	if _, err := b.getTransactions(); err != nil {
		log.Fatalf("error in block gettransactions, %v", err)
		respondWithError(w, http.StatusNotFound, "block not found")
		return
	}
	if _, err := b.getPrice(); err != nil {
		log.Fatalf("error in block gettransactions, %v", err)
		respondWithError(w, http.StatusServiceUnavailable, "prices for block not found")
		return
	}

	respondWithJSON(w, http.StatusOK, b)
}

func (a *App) showAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	address := NewAddress(a.DB)
	err := address.getByHash(vars["hash"])
	if err != nil {
		if err.Error() == "404" {
			respondWithError(w, http.StatusNotFound, "Block not found")
		} else {
			log.Fatalf("app: error in address getbyhash, %v", err)
			respondWithError(w, http.StatusBadRequest, "Cannot retrieve address")
		}
	}

	if _, err := address.getTransactions(); err != nil {
		log.Fatalf("app: error in address gettransactions, %v", err)
	}
	if err := address.getPrices(); err != nil {
		log.Fatalf("app: error in address getprices, %v", err)
	}

	respondWithJSON(w, http.StatusOK, address)
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
