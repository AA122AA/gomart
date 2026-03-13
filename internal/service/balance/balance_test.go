package balance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/repository/mocks"
	"github.com/AA122AA/gomart.git/internal/service/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name       string
		login      string
		setupMocks func(repo *mocks.MockBalanceRepo)
		expCurrent float32
		expSpent   float32
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:  "Posiitve",
			login: "gromartem",
			setupMocks: func(repo *mocks.MockBalanceRepo) {
				res := query.GetBalanceByUserNameRow{
					CurrentBalance:    0.0,
					TotalBonusesSpent: 0.0,
				}
				repo.On("GetByUser", ctx, mock.AnythingOfType("string")).Return(res, nil)
			},
			expCurrent: 0.0,
			expSpent:   0.0,
			pass:       true,
		},
		{
			name:  "repo failed",
			login: "gromartem",
			setupMocks: func(repo *mocks.MockBalanceRepo) {
				repo.On("GetByUser", ctx, mock.AnythingOfType("string")).Return(query.GetBalanceByUserNameRow{}, fmt.Errorf("db connection failed"))
			},
			expCurrent: 0.0,
			expSpent:   0.0,
			pass:       false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection failed")
			},
		},
		{
			name:  "empty username",
			login: "",
			setupMocks: func(repo *mocks.MockBalanceRepo) {
			},
			expCurrent: 0.0,
			expSpent:   0.0,
			pass:       false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrEmptyUsername)
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockBalanceRepo(t)
			tCase.setupMocks(mRepo)

			svc := NewBalanceService(ctx, mRepo)
			res, err := svc.Get(ctx, tCase.login)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, tCase.expCurrent, res.Current)
				require.Equal(t, tCase.expSpent, res.Withdrawn)
				return
			}
			tCase.checkErr(t, err)

		})
	}
}

func TestWithdraw(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name       string
		login      string
		oid        string
		amount     float32
		setupMocks func(repo *mocks.MockBalanceRepo)
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:   "Posiitve",
			login:  "gromartem",
			oid:    "12345678903",
			amount: 1.0,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
				repo.On("WithdrawTransaction", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("int"), mock.AnythingOfType("float32")).Return(nil)
			},
			pass: true,
		},
		{
			name:   "empty username",
			login:  "",
			oid:    "12345678903",
			amount: 1.0,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrEmptyInputData)
			},
		},
		{
			name:   "empty oid",
			login:  "gromartem",
			oid:    "",
			amount: 1.0,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrEmptyInputData)
			},
		},
		{
			name:   "amount 0",
			login:  "gromartem",
			oid:    "12345678903",
			amount: 0,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrEmptyInputData)
			},
		},
		{
			name:   "amount negativ",
			login:  "gromartem",
			oid:    "12345678903",
			amount: -2,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrEmptyInputData)
			},
		},
		{
			name:   "luna error",
			login:  "gromartem",
			oid:    "2345678903",
			amount: 1.0,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				var lerr *utils.ErrWrongLuna
				require.ErrorIs(t, err, lerr)
			},
		},
		{
			name:   "string to int error",
			login:  "gromartem",
			oid:    "lol",
			amount: 1.0,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				var lerr *utils.ErrBadOrderID
				require.ErrorIs(t, err, lerr)
			},
		},
		{
			name:   "db error",
			login:  "gromartem",
			oid:    "12345678903",
			amount: 1.0,
			setupMocks: func(repo *mocks.MockBalanceRepo) {
				repo.On("WithdrawTransaction", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("int"), mock.AnythingOfType("float32")).Return(fmt.Errorf("db connection failed"))
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection failed")
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockBalanceRepo(t)
			tCase.setupMocks(mRepo)

			svc := NewBalanceService(ctx, mRepo)

			err := svc.Withdraw(ctx, tCase.login, tCase.oid, tCase.amount)
			if tCase.pass {
				require.NoError(t, err)
				return
			}
			tCase.checkErr(t, err)

		})
	}
}

func TestHistory(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name       string
		login      string
		oid        string
		setupMocks func(repo *mocks.MockBalanceRepo)
		expBalance float32
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:  "Positive",
			login: "gromartem",
			oid:   "12345678903",
			setupMocks: func(repo *mocks.MockBalanceRepo) {
				res := []query.GetBonusTransactionsByUserNameRow{
					{
						Oid:             12345678903,
						BonusesWithdraw: 20,
						ProcessedAt: pgtype.Timestamptz{
							Time:  time.Now(),
							Valid: true,
						},
					},
				}
				repo.On("GetBonusTransactions", ctx, mock.AnythingOfType("string")).Return(res, nil)
			},
			expBalance: 20.0,
			pass:       true,
		},
		{
			name: "db connection failed",
			setupMocks: func(repo *mocks.MockBalanceRepo) {
				repo.On("GetBonusTransactions", ctx, mock.AnythingOfType("string")).Return(nil, fmt.Errorf("db connection failed"))
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection failed")
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockBalanceRepo(t)
			tCase.setupMocks(mRepo)

			svc := NewBalanceService(ctx, mRepo)

			res, err := svc.History(ctx, tCase.login)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, tCase.oid, res[0].Order)
				require.Equal(t, tCase.expBalance, res[0].Sum)
				return
			}
			tCase.checkErr(t, err)
		})
	}

}
