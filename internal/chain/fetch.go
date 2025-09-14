package chain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ChartResponse struct {
	Values []struct {
		X int64   `json:"x"`
		Y float64 `json:"y"`
	} `json:"values"`
}

type Metrics struct {
	ActiveAddresses int64   `json:"active_addresses"`
	TxCount         int64   `json:"tx_count"`
	Timestamp       int64   `json:"timestamp"`
	Price           float64 `json:"price"`
}

func ActiveAddr() (int64, error) {
	url := "https://api.blockchain.info/charts/n-unique-addresses?timespan=2days&format=json"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch active addresses: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response ChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Values) == 0 {
		return 0, fmt.Errorf("no data returned")
	}

	// Return the latest value
	latest := response.Values[len(response.Values)-1]
	return int64(latest.Y), nil
}

func TxCount() (int64, error) {
	url := "https://api.blockchain.info/charts/n-transactions?timespan=2days&format=json"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch transaction count: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response ChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Values) == 0 {
		return 0, fmt.Errorf("no data returned")
	}

	// Return the latest value
	latest := response.Values[len(response.Values)-1]
	return int64(latest.Y), nil
}

func GetMetrics() (Metrics, error) {
	activeAddr, err := ActiveAddr()
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get active addresses: %w", err)
	}

	txCount, err := TxCount()
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get transaction count: %w", err)
	}

	// Get current BTC price from a free API
	price, err := getBTCPrice()
	if err != nil {
		// If price fetch fails, just log and continue with 0
		price = 0
	}

	return Metrics{
		ActiveAddresses: activeAddr,
		TxCount:         txCount,
		Timestamp:       time.Now().Unix(),
		Price:           price,
	}, nil
}

func getBTCPrice() (float64, error) {
	url := "https://api.coinbase.com/v2/exchange-rates?currency=BTC"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Rates struct {
				USD string `json:"USD"`
			} `json:"rates"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	var price float64
	if _, err := fmt.Sscanf(response.Data.Rates.USD, "%f", &price); err != nil {
		return 0, err
	}

	return price, nil
}
