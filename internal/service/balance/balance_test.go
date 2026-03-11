package balance

import (
	"context"
	"fmt"
	"testing"

	"github.com/AA122AA/gomart.git/internal/db/query"
	"github.com/AA122AA/gomart.git/internal/repository/mocks"
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
