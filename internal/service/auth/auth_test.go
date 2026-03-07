package authservice

import (
	"context"
	"fmt"
	"testing"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/repository/mocks"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	ctx := context.Background()
	login := "gromartem"
	password := "testpass"

	cases := []struct {
		name       string
		setupMocks func(repo *mocks.MockUserRepo)
		cfg        *config.Config
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name: "Positive",
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("CreateUserAndBonusAcc", ctx, login, mock.AnythingOfType("string"), float32(0.0), float32(0.0)).Return(nil)
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: true,
		},
		{
			name: "User Exists",
			setupMocks: func(repo *mocks.MockUserRepo) {
				pgErr := &pgconn.PgError{
					Code: pgerrcode.UniqueViolation,
				}
				repo.On("CreateUserAndBonusAcc", ctx, login, mock.AnythingOfType("string"), float32(0.0), float32(0.0)).Return(pgErr)
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				var uerr *ErrUserExists
				require.ErrorAs(t, err, &uerr)
			},
		},
		{
			name: "DB connection failed",
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("CreateUserAndBonusAcc", ctx, login, mock.AnythingOfType("string"), float32(0.0), float32(0.0)).Return(fmt.Errorf("db connection failed"))
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection failed")
			},
		},
		{
			name: "create bonus acc failed",
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("CreateUserAndBonusAcc", ctx, login, mock.AnythingOfType("string"), float32(0.0), float32(0.0)).Return(fmt.Errorf("can not create bonus acc"))
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "can not create bonus acc")
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockUserRepo(t)
			tCase.setupMocks(mRepo)

			us := NewUserService(ctx, mRepo, tCase.cfg)

			err := us.Register(ctx, login, password)
			if tCase.pass {
				require.NoError(t, err)
				return
			}
			tCase.checkErr(t, err)
		})
	}

}
