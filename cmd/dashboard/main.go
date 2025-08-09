//go:build dashboard

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rivo/tview"
)

// metric mirrors the server's metrics struct.
type metric struct {
	Endpoint string `json:"endpoint"`
	Status   int    `json:"status"`
	Duration string `json:"duration"`
}

func fetchMetrics() ([]metric, error) {
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var m []metric
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func main() {
	app := tview.NewApplication()
	table := tview.NewTable().SetBorders(true)

	update := func() {
		metrics, err := fetchMetrics()
		if err != nil {
			return
		}
		app.QueueUpdateDraw(func() {
			table.Clear()
			headers := []string{"Endpoint", "Status", "Duration"}
			for i, h := range headers {
				table.SetCell(0, i, tview.NewTableCell(h).SetSelectable(false))
			}
			for i, m := range metrics {
				table.SetCell(i+1, 0, tview.NewTableCell(m.Endpoint))
				table.SetCell(i+1, 1, tview.NewTableCell(fmt.Sprintf("%d", m.Status)))
				table.SetCell(i+1, 2, tview.NewTableCell(m.Duration))
			}
		})
	}

	go func() {
		for {
			update()
			time.Sleep(time.Second)
		}
	}()

	if err := app.SetRoot(table, true).Run(); err != nil {
		panic(err)
	}
}
