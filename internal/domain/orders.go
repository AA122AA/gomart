package domain

import (
	"fmt"
	"time"

	"github.com/AA122AA/gomart.git/internal/db/query"
)

type OrderJSON struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func TransformOrderToJSON(in query.GetOrdersByUserNameRow) *OrderJSON {
	return &OrderJSON{
		Number:     fmt.Sprintf("%v", in.Oid),
		Status:     in.Status,
		Accrual:    int(in.Accrual.Int32),
		UploadedAt: in.UploadedAt.Time,
	}
}
