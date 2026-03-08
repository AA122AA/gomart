package authservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/constants"
	"github.com/AA122AA/gomart.git/internal/repository/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	ctx := context.Background()
	login := "gromartem"
	password := "testpass"

	cases := []struct {
		name       string
		login      string
		password   string
		setupMocks func(repo *mocks.MockUserRepo)
		cfg        *config.Config
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:     "Positive",
			login:    login,
			password: password,
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("CreateUserAndBonusAcc", ctx, login, mock.AnythingOfType("string"), float32(0.0), float32(0.0)).Return(nil)
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: true,
		},
		{
			name:     "User Exists",
			login:    login,
			password: password,
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
			name:     "DB connection failed",
			login:    login,
			password: password,
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
			name:     "create bonus acc failed",
			login:    login,
			password: password,
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
		{
			name:     "empty login or password",
			login:    "",
			password: "",
			setupMocks: func(repo *mocks.MockUserRepo) {
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidLoginPassword)
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockUserRepo(t)
			tCase.setupMocks(mRepo)

			us := NewUserService(ctx, mRepo, tCase.cfg)

			err := us.Register(ctx, tCase.login, tCase.password)
			if tCase.pass {
				require.NoError(t, err)
				return
			}
			tCase.checkErr(t, err)
		})
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	login := "gromartem"
	password := "testpass"

	hPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	cases := []struct {
		name       string
		login      string
		password   string
		setupMocks func(repo *mocks.MockUserRepo)
		cfg        *config.Config
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:     "Positive",
			login:    login,
			password: password,
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetPasswordHash", ctx, login).Return(string(hPass), nil)
				repo.On("GetRefreshTokenByUserName", ctx, login).Return(nil, pgx.ErrNoRows)
				repo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), login, true).Return(nil)
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: true,
		},
		{
			name:     "User does not exists",
			login:    login,
			password: password,
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetPasswordHash", ctx, login).Return("", pgx.ErrNoRows)
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidCredentials)
			},
		},
		{
			name:     "DB connection failed",
			login:    login,
			password: password,
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetPasswordHash", ctx, login).Return("", fmt.Errorf("db connection failed"))
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
			name:     "Wrong password",
			login:    login,
			password: "lol",
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetPasswordHash", ctx, login).Return(string(hPass), nil)
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidCredentials)
			},
		},
		{
			name:     "empty login or password",
			login:    "",
			password: "",
			setupMocks: func(repo *mocks.MockUserRepo) {
			},
			cfg: &config.Config{
				Key: "test",
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidLoginPassword)
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockUserRepo(t)
			tCase.setupMocks(mRepo)

			us := NewUserService(ctx, mRepo, tCase.cfg)

			tokens, err := us.Login(ctx, tCase.login, tCase.password)
			if tCase.pass {
				require.NoError(t, err)
				require.NotNil(t, tokens)
				require.NotEmpty(t, tokens.AccessToken)

				claims := &jwt.MapClaims{}
				_, err = jwt.ParseWithClaims(tokens.AccessToken, claims, func(t *jwt.Token) (any, error) {
					return us.key, nil
				})
				require.NoError(t, err)

				require.Equal(t, login, (*claims)["user_name"])
				require.Equal(t, constants.Issuer, (*claims)["iss"])

				if exp, ok := (*claims)["exp"].(float64); ok {
					require.True(t, int64(exp) > time.Now().Unix())
				}

				return
			}
			tCase.checkErr(t, err)
		})
	}
}
