package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Executor struct {
	client  *http.Client
	baseURL string
}

func New(baseURL string) *Executor {
	return &Executor{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 100,
				MaxConnsPerHost:     100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL: baseURL,
	}
}

func (e *Executor) CreateAccount(ctx context.Context, owner string) (string, error) {
	payload := map[string]interface{}{
		"owner": owner,
	}
	
	respBody, err := e.post(ctx, "/accounts", payload)
	if err != nil {
		return "", err
	}
	
	var result struct {
		ID    int    `json:"id"`
		Owner string `json:"owner"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse create account response: %w", err)
	}
	
	return fmt.Sprintf("%d", result.ID), nil
}

func (e *Executor) Deposit(ctx context.Context, accountID string, amount float64) error {
	payload := map[string]int{"amount": int(amount)}
	_, err := e.post(ctx, fmt.Sprintf("/accounts/%s/deposit", accountID), payload)
	return err
}

func (e *Executor) Withdraw(ctx context.Context, accountID string, amount float64) error {
	payload := map[string]int{"amount": int(amount)}
	_, err := e.post(ctx, fmt.Sprintf("/accounts/%s/withdraw", accountID), payload)
	return err
}

func (e *Executor) Transfer(ctx context.Context, fromID, toID string, amount float64) error {
	fromIDInt, err := strconv.Atoi(fromID)
	if err != nil {
		return fmt.Errorf("invalid from account ID: %w", err)
	}
	
	toIDInt, err := strconv.Atoi(toID)
	if err != nil {
		return fmt.Errorf("invalid to account ID: %w", err)
	}
	
	payload := map[string]int{
		"from":   fromIDInt,
		"to":     toIDInt,
		"amount": int(amount),
	}
	_, err = e.post(ctx, "/accounts/transfer", payload)
	return err
}

func (e *Executor) GetBalance(ctx context.Context, accountID string) (float64, error) {
	resp, err := e.get(ctx, fmt.Sprintf("/accounts/%s/balance", accountID))
	if err != nil {
		return 0, err
	}
	
	var result struct {
		Balance float64 `json:"balance"`
	}
	
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("failed to parse balance response: %w", err)
	}
	
	return result.Balance, nil
}

func (e *Executor) post(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Load-Test", "true")
	
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	var respBody bytes.Buffer
	if _, err := respBody.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, respBody.String())
	}
	
	return respBody.Bytes(), nil
}

func (e *Executor) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", e.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Load-Test", "true")
	
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	var respBody bytes.Buffer
	if _, err := respBody.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, respBody.String())
	}
	
	return respBody.Bytes(), nil
}