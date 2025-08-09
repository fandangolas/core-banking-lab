package main

import (
	"bank-api/src/metrics"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var baseURL = getenv("BASE_URL", "http://localhost:8080")

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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

func randomOp(ids []int) {
	switch rand.Intn(3) {
	case 0:
		id := ids[rand.Intn(len(ids))]
		deposit(id, rand.Intn(100)+1)
	case 1:
		id := ids[rand.Intn(len(ids))]
		withdraw(id, rand.Intn(50)+1)
	case 2:
		from := ids[rand.Intn(len(ids))]
		to := ids[rand.Intn(len(ids))]
		for to == from {
			to = ids[rand.Intn(len(ids))]
		}
		transfer(from, to, rand.Intn(30)+1)
	}
}

func main() {
        rand.Seed(time.Now().UnixNano())

        const (
                numAccounts = 100
                totalOps    = 10000
                blockSize   = 100
                blockPause  = 100 * time.Millisecond
        )

        ids := make([]int, 0, numAccounts)
        for i := 0; i < numAccounts; i++ {
                owner := fmt.Sprintf("User%d", i+1)
                id, err := createAccount(owner)
                if err != nil {
                        log.Fatalf("cannot create account %s: %v", owner, err)
                }
                ids = append(ids, id)
                deposit(id, 1000)
        }

        for sent := 0; sent < totalOps; {
                var wg sync.WaitGroup
                for i := 0; i < blockSize && sent < totalOps; i++ {
                        wg.Add(1)
                        go func() {
                                defer wg.Done()
                                randomOp(ids)
                        }()
                        sent++
                }
                wg.Wait()
                time.Sleep(blockPause)
        }

        for _, m := range metrics.List() {
                log.Printf("%s status=%d duration=%s", m.Endpoint, m.Status, m.Duration)
        }
}
