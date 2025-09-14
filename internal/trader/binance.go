package trader

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	apiKey    string
	apiSecret string
	baseURL   string
	client    *http.Client
}

type Order struct {
	Symbol   string `json:"symbol"`
	OrderID  int64  `json:"orderId"`
	ClientID string `json:"clientOrderId"`
	Side     string `json:"side"`
	Type     string `json:"type"`
	Quantity string `json:"origQty"`
	Price    string `json:"price"`
	Status   string `json:"status"`
	Time     int64  `json:"transactTime"`
}

type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   "https://testnet.binance.vision",
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) sign(query string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) TestConnection() error {
	url := c.baseURL + "/api/v3/time"

	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) PlaceOrder(symbol, side string, qty float64) (Order, error) {
	endpoint := "/api/v3/order"

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))
	params.Set("type", "MARKET")
	params.Set("quantity", strconv.FormatFloat(qty, 'f', 8, 64))
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	query := params.Encode()
	signature := c.sign(query)
	query += "&signature=" + signature

	req, err := http.NewRequest("POST", c.baseURL+endpoint+"?"+query, nil)
	if err != nil {
		return Order{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return Order{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return Order{}, fmt.Errorf("API error: %d - %s", errResp.Code, errResp.Msg)
	}

	var order Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return Order{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return order, nil
}

func (c *Client) GetAccountInfo() (map[string]interface{}, error) {
	endpoint := "/api/v3/account"

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	query := params.Encode()
	signature := c.sign(query)
	query += "&signature=" + signature

	req, err := http.NewRequest("GET", c.baseURL+endpoint+"?"+query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
