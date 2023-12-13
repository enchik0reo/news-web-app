package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"newsWebApp/app/authService/internal/config"
	"newsWebApp/app/authService/internal/models"
	"newsWebApp/app/authService/internal/services"
	"newsWebApp/app/authService/internal/services/tokenmanager"
	"newsWebApp/app/authService/internal/storage"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage interface {
	SaveUser(ctx context.Context, userName string, email string, hashPass []byte) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type SessionStorage interface {
	SetSession(ctx context.Context, userID int64, refToken string, expire time.Duration) error
	GetSessionToken(ctx context.Context, userID int64) (string, error)
}

var (
	errCantSaveUser  = errors.New("can't save user")
	errCantLoginUser = errors.New("can't login user")
)

type Auth struct {
	userStorage     UserStorage
	sessionStorage  SessionStorage
	tokenManager    tokenmanager.TokenManager
	log             *slog.Logger
	refreshTokenTTL time.Duration
}

func New(usrS UserStorage, sesS SessionStorage, l *slog.Logger, cfg *config.TokenManager) *Auth {
	tM := tokenmanager.New(cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.SecretKey)

	return &Auth{
		userStorage:     usrS,
		sessionStorage:  sesS,
		tokenManager:    *tM,
		log:             l,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (a *Auth) SaveUser(ctx context.Context, userName string, email string, pass string) (int64, error) {
	if err := validateUser(email, pass); err != nil {
		a.log.Warn("failed get validate user", "err", err.Error())

		return 0, services.ErrInvalidValue
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		a.log.Warn("failed generate hash", "err", err.Error())

		return 0, errCantSaveUser
	}

	id, err := a.userStorage.SaveUser(ctx, userName, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("failed save user", "err", err.Error())

			return 0, services.ErrUserExists
		}
		a.log.Error("failed save user", "err", err.Error())

		return 0, errCantSaveUser
	}

	return id, nil
}

func (a *Auth) LoginUser(ctx context.Context, email, pass string) (int64, string, string, string, error) {
	if err := validateUser(email, pass); err != nil {
		a.log.Warn("failed get validate user", "err", err.Error())

		return 0, "", "", "", services.ErrInvalidValue
	}

	user, err := a.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("failed get user", "err", err.Error())

			return 0, "", "", "", services.ErrUserDoesntExists
		}
		a.log.Error("failed get user", "err", err.Error())

		return 0, "", "", "", errCantLoginUser
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(pass)); err != nil {
		a.log.Warn("failed compare hash and password", "err", err.Error())

		return 0, "", "", "", services.ErrUserDoesntExists
	}

	accToken, refToken, err := a.tokenManager.CreateTokens(user.ID, user.Name)
	if err != nil {
		a.log.Error("failed to create new tokens", "err", err.Error())

		return 0, "", "", "", errCantLoginUser
	}

	if err = a.sessionStorage.SetSession(ctx, user.ID, refToken, a.refreshTokenTTL); err != nil {
		a.log.Error("failed to create session", "err", err.Error())

		return 0, "", "", "", errCantLoginUser
	}

	return user.ID, user.Name, accToken, refToken, nil
}

func (a *Auth) Parse(ctx context.Context, header string) (int64, string, error) {
	token, err := getTokenFromHeader(header)
	if err != nil {
		a.log.Warn(err.Error())

		return 0, "", services.ErrInvalidToken
	}

	id, userName, err := a.tokenManager.Parse(token)
	if err != nil {
		if errors.Is(err, tokenmanager.ErrTokenExpired) {
			a.log.Warn("token expired")

			return 0, "", services.ErrTokenExpired
		}
		a.log.Warn("invalid token", "token", token)

		return 0, "", services.ErrInvalidToken
	}

	return id, userName, nil
}

func (a *Auth) Refresh(ctx context.Context, refrToken string) (int64, string, string, string, error) {
	id, userName, err := a.tokenManager.Parse(refrToken)
	if err != nil {
		if errors.Is(err, tokenmanager.ErrTokenExpired) {
			a.log.Warn("token expired")

			return 0, "", "", "", services.ErrSessionNotFound
		}
		a.log.Warn("invalid token", "token", refrToken)

		return 0, "", "", "", services.ErrInvalidToken
	}

	oldRefreshToken, err := a.sessionStorage.GetSessionToken(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			a.log.Warn("session not found", "user id", id)

			return 0, "", "", "", services.ErrSessionNotFound
		}
		a.log.Error("failed to get user by refresh token", "err", err)

		return 0, "", "", "", errCantLoginUser
	}

	if refrToken == oldRefreshToken {
		newAccToken, newRefToken, err := a.tokenManager.CreateTokens(id, userName)
		if err != nil {
			a.log.Error("failed to create new access token", "err", err.Error())

			return 0, "", "", "", errCantLoginUser
		}

		if err = a.sessionStorage.SetSession(ctx, id, newRefToken, a.refreshTokenTTL); err != nil {
			a.log.Error("failed to create session", "err", err.Error())

			return 0, "", "", "", errCantLoginUser
		}

		return id, userName, newAccToken, newRefToken, nil
	} else {
		a.log.Error("old refresh token != refresh token from user", "old", oldRefreshToken, "now", refrToken)

		return 0, "", "", "", services.ErrInvalidToken
	}
}

func validateUser(email string, pass string) error {
	if err := validator.New(validator.WithRequiredStructEnabled()).
		Var(email, "required,email"); err != nil {
		return err
	}

	if pass == "" {
		return errors.New("password is required")
	}

	return nil
}

func getTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", errors.New("empty auth header")
	}

	headerParts := strings.Split(header, " ")

	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}

	token := headerParts[1]

	if len(token) == 0 {
		return "", errors.New("empty token")
	}

	return token, nil
}
