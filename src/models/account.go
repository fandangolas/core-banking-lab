package models

import (
	"sync"
	"time"
)

type Account struct {
	Id        int       `json:"id"`
	Owner     string    `json:"owner_name"`
	Balance   int       `json:"balance"`
	CreatedAt time.Time `json:"created_at"`

	Mu sync.Mutex `json:"-"`
}
