package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	_ "modernc.org/sqlite"
)

type DollarQuote struct {
	USDBRL struct {
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
	} `json:"USDBRL"`
}

type GetQuoteResult struct {
	Cotacao float64 `json:"cotacao"`
}

type Cotacao struct {
	ID      uint    `gorm:"primaryKey"`
	Cotacao float64 `gorm:"type:REAL"`
}

const quoteApiUrl = "https://economia.awesomeapi.com.br/json/last/"
const externaApiTimeout = 200 * time.Millisecond
const dbTimeout = 10 * time.Millisecond

var (
	ErrTimeout = errors.New("timeout atingido")
	db         *gorm.DB
)

func getQuoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	codeParam := vars["code"]
	codeinParam := vars["codein"]

	quote, err := getQuote(codeParam, codeinParam)

	if err != nil {
		switch err {
		case ErrTimeout:
			http.Error(w, "Timeout ao chamar serviço externo", http.StatusGatewayTimeout)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	errDb := salvarCotacao(quote.Cotacao)

	if errDb != nil {
		http.Error(w, "Erro ao salvar cotação no banco de dados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(quote); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getQuote(code, codein string) (*GetQuoteResult, error) {
	quoteApiUrlWithParams := quoteApiUrl + code + "-" + codein

	ctx, cancel := context.WithTimeout(context.Background(), externaApiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", quoteApiUrlWithParams, nil)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {

		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Timeout atingido após 200ms")
			return nil, ErrTimeout
		}

		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao chamar serviço externo: %s", resp.Status)
	}

	defer resp.Body.Close()

	var dollarQuote DollarQuote

	if err := json.NewDecoder(resp.Body).Decode(&dollarQuote); err != nil {
		return nil, err
	}

	quoteStr := dollarQuote.USDBRL.Bid
	quoteFloat, err := strconv.ParseFloat(quoteStr, 64)

	if err != nil {
		return nil, fmt.Errorf("erro ao converter cotação para float64: %v", err)
	}

	result := GetQuoteResult{Cotacao: quoteFloat}

	return &result, nil
}

func salvarCotacao(valor float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	cotacao := Cotacao{
		Cotacao: valor,
	}

	if err := db.WithContext(ctx).Create(&cotacao).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Timeout ao salvar no banco de dados.")
		}
		return err
	}

	return nil
}

func historicoHandler(w http.ResponseWriter, r *http.Request) {
	var cotacoes []Cotacao
	if err := db.Order("ID desc").Find(&cotacoes).Error; err != nil {
		http.Error(w, "Erro ao buscar histórico", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacoes)
}

func initDB() error {
	var err error

	sqlDB, err := sql.Open("sqlite", "cotacoes.db")

	if err != nil {
		return err
	}

	db, err = gorm.Open(sqlite.Dialector{
		Conn: sqlDB,
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		return err
	}

	return db.AutoMigrate(&Cotacao{})
}

func main() {

	err := initDB()

	if err != nil {
		panic("Erro ao inicializar banco: " + err.Error())
	}

	r := mux.NewRouter()
	r.HandleFunc("/cotacao/{code}/{codein}", getQuoteHandler).Methods("GET")
	r.HandleFunc("/historico", historicoHandler).Methods("GET")
	http.ListenAndServe(":8080", r)
}
