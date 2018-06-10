package vscale

import (
	"log"

	"github.com/vscale/go-vscale"
)

// Account incapsulates Vscale account credential
type Account struct {
	Token  string
	ChatID int64
}

// Balance returns Vscale account balance in roubles
func Balance(token string) float64 {
	client := vscale_api_go.NewClient(token)
	billing, _, err := client.Billing.Billing()
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
	return float64(billing.Balance) / 100
}
