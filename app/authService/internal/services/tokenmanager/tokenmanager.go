package tokenmanager

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token expired")
)

type TokenManager struct {
	accessTTL  time.Duration
	refreshTTL time.Duration
	signingkey string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	RTCreatedAt  time.Time
}

func New(accDur time.Duration, refrDur time.Duration, skey string) *TokenManager {
	return &TokenManager{
		accessTTL:  accDur,
		refreshTTL: refrDur,
		signingkey: skey,
	}
}

func (t *TokenManager) CreateTokens(userID int64, userName string) (string, string, error) {
	acsT, err := t.newJWT(userID, userName)
	if err != nil {
		return "", "", err
	}

	rfrT, err := t.newRefreshToken(userID, userName)
	if err != nil {
		return "", "", err
	}

	return acsT, rfrT, nil
}

func (t *TokenManager) Parse(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(t.signingkey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, "", ErrTokenExpired
		}
		return 0, "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fmt.Errorf("error get user claims from token")
	}

	fl64 := claims["sub"].(float64)
	it64 := int64(fl64)

	userName := claims["unm"].(string)

	return it64, userName, nil
}

func (t *TokenManager) newJWT(userID int64, userName string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["unm"] = userName
	claims["exp"] = time.Now().Add(t.accessTTL).UTC().Unix()

	return token.SignedString([]byte(t.signingkey))
}

func (t *TokenManager) newRefreshToken(userID int64, userName string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["unm"] = userName
	claims["exp"] = time.Now().Add(t.refreshTTL).UTC().Unix()

	return token.SignedString([]byte(t.signingkey))
}
