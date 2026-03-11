package authservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/constants"
	"github.com/AA122AA/gomart.git/internal/domain"
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

func TestLogout(t *testing.T) {
	ctx := context.Background()

	login := "gromartem"
	mRepo := mocks.NewMockUserRepo(t)
	cfg := &config.Config{
		Key: "test",
	}

	us := NewUserService(ctx, mRepo, cfg)

	cases := []struct {
		name        string
		accessToken string
		setupMocks  func(repo *mocks.MockUserRepo)
		pass        bool
		checkErr    func(t *testing.T, err error)
	}{
		{
			name: "Positive",
			accessToken: func() string {
				token, err := us.createAccessToken(login)
				require.NoError(t, err)
				return token
			}(),
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("InvalidateRefreshToken", ctx, mock.AnythingOfType("string")).Return(nil)
			},
			pass: true,
		},
		{
			name:        "Invalid token",
			accessToken: "alb",
			setupMocks: func(repo *mocks.MockUserRepo) {
				// repo.On("InvalidateRefreshToken", ctx, mock.AnythingOfType("string")).Return(nil)
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidToken)
			},
		},
		{
			name: "DB connection failed",
			accessToken: func() string {
				token, err := us.createAccessToken(login)
				require.NoError(t, err)
				return token
			}(),
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("InvalidateRefreshToken", ctx, mock.AnythingOfType("string")).Return(fmt.Errorf("db connection failed"))
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection failed")
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo.ExpectedCalls = nil
			tCase.setupMocks(mRepo)

			err := us.Logout(ctx, tCase.accessToken)
			if tCase.pass {
				require.NoError(t, err)

				return
			}
			tCase.checkErr(t, err)
		})
	}
}

func TestRefresh(t *testing.T) {
	ctx := context.Background()

	login := "gromartem"
	mRepo := mocks.NewMockUserRepo(t)
	cfg := &config.Config{
		Key: "test",
	}
	mRepo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, pgx.ErrNoRows)
	mRepo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)

	us := NewUserService(ctx, mRepo, cfg)

	token, err := us.createRefreshToken(ctx, login)
	require.NoError(t, err)

	mRepo.ExpectedCalls = nil

	cases := []struct {
		name         string
		refreshToken string
		setupMocks   func(repo *mocks.MockUserRepo)
		pass         bool
		checkErr     func(t *testing.T, err error)
	}{
		{
			name:         "Positive",
			refreshToken: token,
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(&domain.GetRefreshTokenID{IsValid: true}, nil).Once()
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, pgx.ErrNoRows)
				repo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)
			},
			pass: true,
		},
		{
			name:         "Invalid token",
			refreshToken: "alb",
			setupMocks: func(repo *mocks.MockUserRepo) {
				// repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(&domain.GetRefreshTokenID{IsValid: true}, nil)
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidToken)
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo.ExpectedCalls = nil
			tCase.setupMocks(mRepo)

			tokens, err := us.Refresh(ctx, tCase.refreshToken)
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

func TestCreateTokens(t *testing.T) {
	ctx := context.Background()

	login := "gromartem"

	cases := []struct {
		name       string
		login      string
		setupMocks func(repo *mocks.MockUserRepo)
		cfg        *config.Config
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:  "Positive",
			login: login,
			cfg: &config.Config{
				Key: "test",
			},
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, pgx.ErrNoRows)
				repo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)
			},
			pass: true,
		},
		{
			name:  "cannot create refresh token",
			login: "alb",
			cfg: &config.Config{
				Key: "test",
			},
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, fmt.Errorf("db connection error"))
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection error")
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockUserRepo(t)

			us := NewUserService(ctx, mRepo, tCase.cfg)
			mRepo.ExpectedCalls = nil
			tCase.setupMocks(mRepo)

			tokens, err := us.createTokens(ctx, tCase.login)
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

				refreshClaims := jwt.MapClaims{}
				refreshToken, err := jwt.ParseWithClaims(tokens.RefreshToken, refreshClaims, func(token *jwt.Token) (interface{}, error) {
					return []byte(tCase.cfg.Key), nil
				})
				require.NoError(t, err)
				require.True(t, refreshToken.Valid)

				return
			}
			tCase.checkErr(t, err)
		})
	}
}

func TestCreateRefreshToken(t *testing.T) {
	ctx := context.Background()

	login := "gromartem"

	cases := []struct {
		name       string
		login      string
		setupMocks func(repo *mocks.MockUserRepo)
		cfg        *config.Config
		pass       bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:  "Positive create",
			login: login,
			cfg: &config.Config{
				Key: "test",
			},
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, pgx.ErrNoRows)
				repo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)
			},
			pass: true,
		},
		{
			name:  "cannot update token",
			login: login,
			cfg: &config.Config{
				Key: "test",
			},
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, nil)
				repo.On("UpdateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)
			},
			pass: true,
		},
		{
			name:  "cannot get token, db error",
			login: "alb",
			cfg: &config.Config{
				Key: "test",
			},
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, fmt.Errorf("db connection error"))
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection error")
			},
		},
		{
			name:  "cannot update token",
			login: login,
			cfg: &config.Config{
				Key: "test",
			},
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, nil)
				repo.On("UpdateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(fmt.Errorf("db connection error"))
				// repo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection error")
			},
		},
		{
			name:  "cannot create token",
			login: login,
			cfg: &config.Config{
				Key: "test",
			},
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, pgx.ErrNoRows)
				repo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(fmt.Errorf("db connection error"))
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection error")
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo := mocks.NewMockUserRepo(t)

			us := NewUserService(ctx, mRepo, tCase.cfg)
			mRepo.ExpectedCalls = nil
			tCase.setupMocks(mRepo)

			refreshToken, err := us.createRefreshToken(ctx, tCase.login)
			if tCase.pass {
				require.NoError(t, err)
				require.NotNil(t, refreshToken)
				require.NotEmpty(t, refreshToken)

				claims := &jwt.MapClaims{}
				rToken, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (any, error) {
					return us.key, nil
				})
				require.NoError(t, err)

				require.True(t, rToken.Valid)
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

func TestVerifyToken(t *testing.T) {
	ctx := context.Background()

	login := "gromartem"
	mRepo := mocks.NewMockUserRepo(t)
	cfg := &config.Config{
		Key: "test",
	}
	mRepo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, pgx.ErrNoRows)
	mRepo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)

	us := NewUserService(ctx, mRepo, cfg)

	tokens, err := us.createTokens(ctx, login)
	require.NoError(t, err)

	mRepo.ExpectedCalls = nil

	cases := []struct {
		name     string
		token    string
		pass     bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:  "Positive access",
			token: tokens.AccessToken,
			pass:  true,
		},
		{
			name:  "Positive refresh",
			token: tokens.RefreshToken,
			pass:  true,
		},
		{
			name:  "Invalid token parse",
			token: "ljsdflj",
			pass:  false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, jwt.ErrTokenMalformed)
			},
		},
		{
			name: "Invalid sign",
			token: func() string {
				now := time.Now()
				var accessClaims = jwt.MapClaims{
					"iss":       constants.Issuer,
					"sub":       login,
					"iat":       now.Unix(),
					"exp":       now.Add(15 * time.Minute).Unix(),
					"user_name": login,
				}

				accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
				token, err := accToken.SignedString([]byte("lol"))
				require.NoError(t, err)
				return token
			}(),
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, jwt.ErrSignatureInvalid)
			},
		},
		{
			name: "Invalid issuer",
			token: func() string {
				now := time.Now()
				var accessClaims = jwt.MapClaims{
					"iss":       "lol",
					"sub":       login,
					"iat":       now.Unix(),
					"exp":       now.Add(15 * time.Minute).Unix(),
					"user_name": login,
				}

				accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
				token, err := accToken.SignedString([]byte(cfg.Key))
				require.NoError(t, err)
				return token
			}(),
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, jwt.ErrTokenInvalidIssuer)
			},
		},
		{
			name: "Invalid sign method",
			token: func() string {
				now := time.Now()
				var accessClaims = jwt.MapClaims{
					"iss":       constants.Issuer,
					"sub":       login,
					"iat":       now.Unix(),
					"exp":       now.Add(15 * time.Minute).Unix(),
					"user_name": login,
				}

				accToken := jwt.NewWithClaims(jwt.SigningMethodNone, accessClaims)
				token, err := accToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
				require.NoError(t, err)
				return token
			}(),
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
			},
		},
		{
			name: "invalid username",
			token: func() string {
				now := time.Now()
				var accessClaims = jwt.MapClaims{
					"iss":       constants.Issuer,
					"sub":       login,
					"iat":       now.Unix(),
					"exp":       now.Add(15 * time.Minute).Unix(),
					"user_name": 1,
				}

				accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
				token, err := accToken.SignedString([]byte(cfg.Key))
				require.NoError(t, err)
				return token
			}(),
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidToken)
			},
		},
		{
			name: "invalid mapclaims",
			token: func() string {
				now := time.Now()
				type CustomClaims struct {
					Lol string `json:"lol"`
					jwt.RegisteredClaims
				}
				accessClaims := CustomClaims{
					Lol: "lol",
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    constants.Issuer,
						Subject:   login,
						IssuedAt:  jwt.NewNumericDate(now),
						ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
					},
				}

				accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
				token, err := accToken.SignedString([]byte(cfg.Key))
				require.NoError(t, err)
				return token
			}(),
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidToken)
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			user, err := us.verifyToken(tCase.token)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, login, user)
				return
			}
			tCase.checkErr(t, err)
		})
	}
}

func TestVerifyRefreshToken(t *testing.T) {
	ctx := context.Background()

	login := "gromartem"
	mRepo := mocks.NewMockUserRepo(t)
	cfg := &config.Config{
		Key: "test",
	}
	mRepo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, pgx.ErrNoRows)
	mRepo.On("CreateRefreshToken", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil)

	us := NewUserService(ctx, mRepo, cfg)

	token, err := us.createRefreshToken(ctx, login)
	require.NoError(t, err)

	mRepo.ExpectedCalls = nil

	cases := []struct {
		name         string
		refreshToken string
		setupMocks   func(repo *mocks.MockUserRepo)
		pass         bool
		checkErr     func(t *testing.T, err error)
	}{
		{
			name:         "Positive",
			refreshToken: token,
			setupMocks: func(repo *mocks.MockUserRepo) {
				tokenInfo := &domain.GetRefreshTokenID{
					TokenID: "lol",
					IsValid: true,
				}
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(tokenInfo, nil)
			},
			pass: true,
		},
		{
			name:         "bad token",
			refreshToken: "alb",
			setupMocks: func(repo *mocks.MockUserRepo) {
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "parse")
			},
		},
		{
			name:         "repo error",
			refreshToken: token,
			setupMocks: func(repo *mocks.MockUserRepo) {
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(nil, fmt.Errorf("db connection failed"))
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "db connection failed")
			},
		},
		{
			name:         "invalid token",
			refreshToken: token,
			setupMocks: func(repo *mocks.MockUserRepo) {
				tokenInfo := &domain.GetRefreshTokenID{
					TokenID: "lol",
					IsValid: false,
				}
				repo.On("GetRefreshTokenByUserName", ctx, mock.AnythingOfType("string")).Return(tokenInfo, nil)
			},
			pass: false,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, ErrInvalidToken)
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mRepo.ExpectedCalls = nil
			tCase.setupMocks(mRepo)

			user, err := us.verifyRefreshToken(ctx, tCase.refreshToken)
			if tCase.pass {
				require.NoError(t, err)
				require.Equal(t, login, user)
				return
			}
			tCase.checkErr(t, err)
		})
	}
}
