package domain

import (
	"fmt"
	"time"

	"github.com/AA122AA/gomart.git/internal/db/query"
)

type BalanceJSON struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type WithdrawJSON struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type HistrotyJSON struct {
	History []*WithdrawJSON
}

func TransformHistoryToJSON(b query.GetBonusTransactionsByUserNameRow) *WithdrawJSON {
	return &WithdrawJSON{
		Order:       fmt.Sprintf("%d", b.Oid),
		Sum:         b.BonusesWithdraw,
		ProcessedAt: b.ProcessedAt.Time,
	}
}

func TransformBalanceToJSON(b query.GetBalanceByUserNameRow) *BalanceJSON {
	return &BalanceJSON{
		Current:   b.CurrentBalance,
		Withdrawn: b.TotalBonusesSpent,
	}
}
