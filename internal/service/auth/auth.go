package authservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AA122AA/gomart.git/internal/config"
	"github.com/AA122AA/gomart.git/internal/constants"
	"github.com/AA122AA/gomart.git/internal/domain"
	"github.com/go-faster/sdk/zctx"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUserAndBonusAcc(ctx context.Context, username, passwordHash string, balance, spent float32) error
	GetPasswordHash(ctx context.Context, username string) (string, error)
	GetRefreshTokenByUserName(ctx context.Context, username string) (*domain.GetRefreshTokenID, error)
	CreateRefreshToken(ctx context.Context, tokenID, username string, isvalid bool) error
	UpdateRefreshToken(ctx context.Context, tokenID, username string, isvalid bool) error
	InvalidateRefreshToken(ctx context.Context, username string) error
}

type UserService struct {
	repo UserRepo

	key []byte
	lg  *zap.Logger
}

func NewUserService(ctx context.Context, repo UserRepo, cfg *config.Config) *UserService {
	return &UserService{
		repo: repo,
		key:  []byte(cfg.Key),
		lg:   zctx.From(ctx).Named("auth service"),
	}
}

func (as *UserService) Register(ctx context.Context, login, password string) error {
	if login == "" || password == "" {
		return ErrInvalidLoginPassword
	}

	hPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("cannot generate hash from password: %w", err)
	}

	err = as.repo.CreateUserAndBonusAcc(ctx, login, string(hPass), float32(0.0), float32(0.0))
	if err != nil {
		var perr *pgconn.PgError
		if errors.As(err, &perr) && perr.Code == pgerrcode.UniqueViolation {
			return NewErrUserExists(err)
		}
		return err
	}

	return nil
}

func (as *UserService) Login(ctx context.Context, login, password string) (*domain.AccessRefreshJSON, error) {
	if login == "" || password == "" {
		return nil, ErrInvalidLoginPassword
	}

	hPass, err := as.repo.GetPasswordHash(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			as.lg.Error("login for not existing user", zap.String("login", login))
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hPass), []byte(password))
	if err != nil {
		as.lg.Error("invalid password", zap.String("login", login))
		return nil, ErrInvalidCredentials
	}

	return as.createTokens(ctx, login)
}

func (as *UserService) Logout(ctx context.Context, accessToken string) error {
	userName, err := as.verifyToken(accessToken)
	if err != nil {
		return fmt.Errorf("cannot verify access token: %w", err)
	}

	err = as.repo.InvalidateRefreshToken(ctx, userName)
	if err != nil {
		return fmt.Errorf("cannot delete refresh token: %w", err)
	}

	return nil
}

func (as *UserService) Refresh(ctx context.Context, token string) (*domain.AccessRefreshJSON, error) {
	user, err := as.verifyRefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return as.createTokens(ctx, user)
}

func (as *UserService) GetUserByToken(ctx context.Context, rawToken string) (string, error) {
	return as.verifyToken(rawToken)
}

func (as *UserService) VerifyAccessToken(ctx context.Context, rawToken string) error {
	_, err := as.verifyToken(rawToken)

	return err
}
func (as *UserService) createTokens(ctx context.Context, login string) (*domain.AccessRefreshJSON, error) {
	accessToken, err := as.createAccessToken(login)
	if err != nil {
		return nil, err
	}

	refreshToken, err := as.createRefreshToken(ctx, login)
	if err != nil {
		return nil, err
	}

	return &domain.AccessRefreshJSON{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, err
}

func (as *UserService) createAccessToken(login string) (string, error) {
	now := time.Now()
	var accessClaims = jwt.MapClaims{
		"iss":       constants.Issuer,
		"sub":       login,
		"iat":       now.Unix(),
		"exp":       now.Add(15 * time.Minute).Unix(),
		"user_name": login,
	}

	accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	return accToken.SignedString(as.key)
}

func (as *UserService) createRefreshToken(ctx context.Context, login string) (string, error) {
	tokenID := uuid.New().String()
	now := time.Now()
	refreshClaims := jwt.MapClaims{
		"iss":       constants.Issuer, // кто выдал токен
		"sub":       login,
		"iat":       now.Unix(),
		"exp":       now.Add(7 * 24 * time.Hour).Unix(),
		"jti":       tokenID,
		"user_name": login,

		"type": "refresh",
	}

	refToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshToken, err := refToken.SignedString(as.key)
	if err != nil {
		return "", err
	}

	_, err = as.repo.GetRefreshTokenByUserName(ctx, login)
	if err == nil {
		err = as.repo.UpdateRefreshToken(ctx, tokenID, login, true)
		if err != nil {
			return "", err
		}

		return refreshToken, nil
	}

	// if strings.Contains(err.Error(), "no rows in result set") {
	if errors.Is(err, pgx.ErrNoRows) {
		err = as.repo.CreateRefreshToken(ctx, tokenID, login, true)
		if err != nil {
			return "", err
		}

		return refreshToken, nil
	}

	return "", err
}

func (as *UserService) keyFunc() jwt.Keyfunc {
	return func(_ *jwt.Token) (interface{}, error) { return as.key, nil }
}

func (as *UserService) verifyToken(rawToken string) (string, error) {
	token, err := jwt.Parse(
		rawToken,
		// получаем ключ для проверки подписи
		as.keyFunc(),
		// проверяем соответсвие алгоритма подписи
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		// проверяем соответствие автора токена
		jwt.WithIssuer(constants.Issuer),
		// проверяем время жизни токена
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("parse token failed: %w", err)
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	userName, ok := claims["user_name"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userName, nil
}

func (as *UserService) verifyRefreshToken(ctx context.Context, rawToken string) (string, error) {
	user, err := as.verifyToken(rawToken)
	if err != nil {
		return "", err
	}

	tokenInfo, err := as.repo.GetRefreshTokenByUserName(ctx, user)
	if err != nil {
		as.lg.Debug("no token for user", zap.String("user", user), zap.String("token", rawToken), zap.Error(err))
		return "", err
	}

	if !tokenInfo.IsValid {
		return "", ErrInvalidToken
	}

	return user, nil
}
