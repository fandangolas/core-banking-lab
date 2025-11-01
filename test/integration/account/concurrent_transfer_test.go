package account

import (
	"bank-api/internal/infrastructure/database"
	"bank-api/test/integration/testenv"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrentTransfer(t *testing.T) {
	router := testenv.SetupRouter()
	defer database.Repo.Reset()

	fromID := testenv.CreateAccount(t, router, "Fonte")
	toID := testenv.CreateAccount(t, router, "Destino")

	// Damos saldo inicial à conta origem
	testenv.Deposit(t, router, fromID, 10000) // R$ 100,00

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

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Errorf("Erro na transferência: %d", resp.Code)
			}
		}()
	}

	wg.Wait()

	fromFinal := testenv.GetBalance(t, router, fromID)
	toFinal := testenv.GetBalance(t, router, toID)
	expected := n * amount

	require.Equal(t, 10000-expected, fromFinal)
	require.Equal(t, expected, toFinal)
}
