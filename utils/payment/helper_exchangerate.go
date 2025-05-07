package utilspayment

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type FXResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

func GetFXRateTHBtoUSD() (float64, error) {
	resp, err := http.Get("https://open.er-api.com/v6/latest/THB")
	if err != nil {
		return 0, fmt.Errorf("failed to call FX API: %v", err)
	}
	defer resp.Body.Close()

	var fxResp FXResponse
	if err := json.NewDecoder(resp.Body).Decode(&fxResp); err != nil {
		return 0, fmt.Errorf("failed to decode FX response: %v", err)
	}

	if fxResp.Result != "success" {
		return 0, fmt.Errorf("API returned failure result")
	}

	rate, ok := fxResp.Rates["USD"]
	if !ok {
		return 0, fmt.Errorf("USD rate not found")
	}

	return rate, nil
}
