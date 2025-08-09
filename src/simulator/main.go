package main

import (
	"bank-api/src/metrics"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const baseURL = "http://localhost:8080"

func createAccount(owner string) (int, error) {
	body, _ := json.Marshal(map[string]string{"owner": owner})
	start := time.Now()
	resp, err := http.Post(baseURL+"/accounts", "application/json", bytes.NewReader(body))
	duration := time.Since(start)
	status := 0
	if err != nil {
		metrics.Record("/accounts", status, duration)
		return 0, err
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	metrics.Record("/accounts", status, duration)
	var data struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	return data.ID, nil
}

func deposit(id, amount int) {
	endpoint := fmt.Sprintf("/accounts/%d/deposit", id)
	body, _ := json.Marshal(map[string]int{"amount": amount})
	start := time.Now()
	resp, err := http.Post(baseURL+endpoint, "application/json", bytes.NewReader(body))
	duration := time.Since(start)
	status := 0
	if err == nil {
		status = resp.StatusCode
		resp.Body.Close()
	} else {
		log.Printf("deposit error: %v", err)
	}
	metrics.Record(endpoint, status, duration)
}

func withdraw(id, amount int) {
	endpoint := fmt.Sprintf("/accounts/%d/withdraw", id)
	body, _ := json.Marshal(map[string]int{"amount": amount})
	start := time.Now()
	resp, err := http.Post(baseURL+endpoint, "application/json", bytes.NewReader(body))
	duration := time.Since(start)
	status := 0
	if err == nil {
		status = resp.StatusCode
		resp.Body.Close()
	} else {
		log.Printf("withdraw error: %v", err)
	}
	metrics.Record(endpoint, status, duration)
}

func transfer(from, to, amount int) {
	endpoint := "/accounts/transfer"
	body, _ := json.Marshal(map[string]int{"from": from, "to": to, "amount": amount})
	start := time.Now()
	resp, err := http.Post(baseURL+endpoint, "application/json", bytes.NewReader(body))
	duration := time.Since(start)
	status := 0
	if err == nil {
		status = resp.StatusCode
		resp.Body.Close()
	} else {
		log.Printf("transfer error: %v", err)
	}
	metrics.Record(endpoint, status, duration)
}

func main() {
	ids := make([]int, 0, 2)
	for _, owner := range []string{"Alice", "Bob"} {
		id, err := createAccount(owner)
		if err != nil {
			log.Fatalf("cannot create account %s: %v", owner, err)
		}
		ids = append(ids, id)
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		deposit(ids[0], 100)
	}()
	go func() {
		defer wg.Done()
		withdraw(ids[0], 50)
	}()
	go func() {
		defer wg.Done()
		transfer(ids[0], ids[1], 25)
	}()
	wg.Wait()

	for _, m := range metrics.List() {
		log.Printf("%s status=%d duration=%s", m.Endpoint, m.Status, m.Duration)
	}
}
