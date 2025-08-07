package testenv

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func CreateAccount(t *testing.T, owner string) int {
	body := map[string]interface{}{"owner": owner}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	SetupRouter().ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("erro ao criar conta: %d", resp.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	return int(result["id"].(float64))
}

func GetBalance(t *testing.T, id int) int {
	req := httptest.NewRequest("GET", "/accounts/"+strconv.Itoa(id)+"/balance", nil)
	resp := httptest.NewRecorder()

	SetupRouter().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("erro ao consultar saldo: %d", resp.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	return int(result["balance"].(float64))
}

func Deposit(t *testing.T, id int, amount int) {
	body := map[string]int{"amount": amount}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/accounts/"+strconv.Itoa(id)+"/deposit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	SetupRouter().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("erro no depósito: %d", resp.Code)
	}
}
