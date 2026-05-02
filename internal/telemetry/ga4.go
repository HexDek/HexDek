package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type GA4Client struct {
	MeasurementID string
	APISecret     string
	ClientID      string
	httpClient    *http.Client
}

func NewGA4Client() *GA4Client {
	mid := os.Getenv("HEXDEK_GA4_MEASUREMENT_ID")
	secret := os.Getenv("HEXDEK_GA4_API_SECRET")
	if mid == "" || secret == "" {
		return nil
	}
	return &GA4Client{
		MeasurementID: mid,
		APISecret:     secret,
		ClientID:      "hexdek-engine-1",
		httpClient:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *GA4Client) SendEvent(name string, params map[string]interface{}) {
	if c == nil {
		return
	}
	payload := map[string]interface{}{
		"client_id": c.ClientID,
		"events": []map[string]interface{}{
			{
				"name":   name,
				"params": params,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	url := "https://www.google-analytics.com/mp/collect?measurement_id=" + c.MeasurementID + "&api_secret=" + c.APISecret
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	resp.Body.Close()
}
