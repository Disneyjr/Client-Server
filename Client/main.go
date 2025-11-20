package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type BidCurrencyConvertUSDBRLResponse struct {
	Bid string `json:"bid"`
}

func NewBidCurrencyConvertUSDBRLResponse() *BidCurrencyConvertUSDBRLResponse {
	return &BidCurrencyConvertUSDBRLResponse{}
}
func main() {
	fmt.Println("Starting the Client.")
	file, err := os.OpenFile("./cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create("./cotacao.txt")
			if err != nil {
				fmt.Println("Não foi possivel criar o arquivo.")
			}
		}
	}
	defer file.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
		fmt.Println("O Servidor demorou demais para responder" + ctx.Err().Error())
		return
	default:
		fmt.Println("Start Req")
		defer fmt.Println("End Req")
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

		if err != nil {
			fmt.Println("Erro ao realizar a requisicação")
			fmt.Println(err.Error())
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Erro ao realizar a requisicação")
			fmt.Println(err.Error())
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Erro ao realizar a requisicação")
			fmt.Println(err.Error())
		}
		fmt.Println(string(body))
		bidResp := NewBidCurrencyConvertUSDBRLResponse()
		err = json.Unmarshal(body, &bidResp)
		if err != nil {
			fmt.Println("Erro no parse do JSON " + err.Error())
		}
		if file != nil {
			if _, err := file.WriteString("Dólar: " + bidResp.Bid + "\n"); err != nil {
				fmt.Println("Erro ao escrever no arquivo:", err)
				return
			}
		}
	}
}
