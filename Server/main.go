package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const urlAwesomeAPI = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type CurrencyConvertResponseUSDBRL struct {
	CurrencyConvertUSDBRL CurrencyConvertUSDBRL `json:"USDBRL"`
}
type CurrencyConvertUSDBRL struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}
type BidCurrencyConvertUSDBRLResponse struct {
	Bid string `json:"bid"`
}
type Server struct {
	db *sql.DB
}

func NewBidCurrencyConvertUSDBRLResponse(bid string) *BidCurrencyConvertUSDBRLResponse {
	return &BidCurrencyConvertUSDBRLResponse{Bid: bid}
}
func NewCurrencyConvertResponseUSDBRL() *CurrencyConvertResponseUSDBRL {
	return &CurrencyConvertResponseUSDBRL{}
}
func NewCurrencyConvertUSDBRL() *CurrencyConvertUSDBRL {
	return &CurrencyConvertUSDBRL{}
}
func NewServer(db *sql.DB) *Server {
	return &Server{
		db: db,
	}
}

func main() {

	log.Println("Starting the server.")

	dbPath := "./data/currency.db"
	dir := filepath.Dir(dbPath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Erro criando diretório: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Creating a DB")
	CreateTables(db)
	server := NewServer(db)
	http.HandleFunc("/cotacao", server.CurrencyHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Hellou!")) })

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Erro ao criar o server.")
	}
}

func (s *Server) CurrencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawPath != "/cotacao" {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	log.Println("Start Req")
	defer log.Println("End Req")
	data := NewCurrencyConvertResponseUSDBRL()
	ctxAPI, cancelShort := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancelShort()
	select {
	case <-ctxAPI.Done():
		http.Error(w, "A API demorou muito para responder, "+ctxAPI.Err().Error(), http.StatusInternalServerError)
		return
	default:
		req, err := http.NewRequestWithContext(ctxAPI, http.MethodGet, urlAwesomeAPI, nil)
		if err != nil {
			http.Error(w, "Erro ao criar request "+err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Erro ao realizar a request "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		rbody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Erro ao ler resposta "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(rbody, &data); err != nil {
			http.Error(w, "Erro no parse do JSON "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	ctxDB, cancelLong := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancelLong()
	select {
	case <-ctxDB.Done():
		http.Error(w, "O DB demorou muito para responder, "+ctxAPI.Err().Error(), http.StatusInternalServerError)
		return
	default:
		insertSQL := `INSERT INTO currency (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date)
									VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
		_, err := s.db.ExecContext(ctxDB, insertSQL,
			data.CurrencyConvertUSDBRL.Code,
			data.CurrencyConvertUSDBRL.Codein,
			data.CurrencyConvertUSDBRL.Name,
			data.CurrencyConvertUSDBRL.High,
			data.CurrencyConvertUSDBRL.Low,
			data.CurrencyConvertUSDBRL.VarBid,
			data.CurrencyConvertUSDBRL.PctChange,
			data.CurrencyConvertUSDBRL.Bid,
			data.CurrencyConvertUSDBRL.Ask,
			data.CurrencyConvertUSDBRL.Timestamp,
			data.CurrencyConvertUSDBRL.CreateDate,
		)
		if err != nil {
			log.Fatalln("Erro no banco de dados " + err.Error())
			http.Error(w, "Erro no banco de dados "+err.Error(), http.StatusInternalServerError)
			return
		}
		responseBid := NewBidCurrencyConvertUSDBRLResponse(data.CurrencyConvertUSDBRL.Bid)
		if err := json.NewEncoder(w).Encode(responseBid); err != nil {
			log.Fatalln("Erro ao serializar " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func CreateTables(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS currency (
    code TEXT,
    codein TEXT,
    name TEXT,
    high TEXT,
    low TEXT,
    varBid TEXT,
    pctChange TEXT,
    bid TEXT,
    ask TEXT,
    timestamp TEXT,
    create_date TEXT
);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Erro criando tabela currency: %s\n", err)
	}
}
