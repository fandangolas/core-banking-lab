package account

import (
	"bank-api/src/db"
	"bank-api/tests/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestConcurrentTransfer(t *testing.T) {
	testenv.SetupRouter()
	defer db.InMemory.Reset()

	fromID := testenv.CreateAccount(t, "Fonte")
	toID := testenv.CreateAccount(t, "Destino")

	// Damos saldo inicial à conta origem
	testenv.Deposit(t, fromID, 10000) // R$ 100,00

	var wg sync.WaitGroup
	n := 100
	amount := 100 // R$ 1,00 por transferência
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			body := map[string]int{
				"from":   fromID,
				"to":     toID,
				"amount": amount,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/accounts/transfer", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			testenv.SetupRouter().ServeHTTP(resp, req)
		}()
	}

	wg.Wait()

	fromFinal := testenv.GetBalance(t, fromID)
	toFinal := testenv.GetBalance(t, toID)
	expected := n * amount

	if fromFinal != (10000-expected) || toFinal != expected {
		t.Errorf("CONCORRÊNCIA DETECTADA: Saldo incorreto.\n  Origem: %d (esperado %d)\n  Destino: %d (esperado %d)",
			fromFinal, 10000-expected, toFinal, expected)
	} else {
		t.Logf("Sem problemas detectados — mas improvável sem mutex!")
	}
}
