package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"newsWebApp/app/apiService/internal/services"
)

type signUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func signup(timeout time.Duration, service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := signUpRequest{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from sign-up request", "err", err.Error())

				err := responseJSONError(w, http.StatusBadRequest, 0, "", "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := service.SaveUser(ctx, req.Name, req.Email, req.Password); err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				err = responseJSONError(w, http.StatusNoContent, 0, "", "User already exists")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrInvalidValue):
				err = responseJSONError(w, http.StatusBadRequest, 0, "", "Invalid value")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't save user", "err", err.Error())

				err = responseJSONError(w, http.StatusInternalServerError, 0, "", "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		err := responseJSON(w, http.StatusCreated, 0, "", nil)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func login(timeout time.Duration, refTokTTL time.Duration, service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := loginRequest{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from login request", "err", err.Error())

				err = responseJSONError(w, http.StatusBadRequest, 0, "", "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		_, _, acsToken, refToken, err := service.LoginUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				err = responseJSONError(w, http.StatusNoContent, 0, "", "Wrong e-mail or password")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			case errors.Is(err, services.ErrInvalidValue):
				err = responseJSONError(w, http.StatusBadRequest, 0, "", "Invalid value")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			default:
				slog.Error("Can't login user", "err", err.Error())

				err = responseJSONError(w, http.StatusInternalServerError, 0, "", "Internal error")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ck := http.Cookie{
			Name:     "refresh_token",
			Domain:   r.URL.Host,
			Path:     "/",
			Value:    refToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(refTokTTL),
		}

		http.SetCookie(w, &ck)

		err = responseJSON(w, http.StatusAccepted, 0, acsToken, nil)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type checkEmailRequest struct {
	Email string `json:"email"`
}

func checkEmail(timeout time.Duration, service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := checkEmailRequest{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from sign-up request", "err", err.Error())

				err := responseJSONError(w, http.StatusBadRequest, 0, "", "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		isExists, err := service.CheckEmail(ctx, req.Email)
		if err != nil {
			slog.Error("Can't check e-mail", "err", err.Error())

			err = responseJSONError(w, http.StatusInternalServerError, 0, "", "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		err = responseCheckJSON(w, http.StatusOK, isExists)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}

type checkUserNameRequest struct {
	UserName string `json:"user_name"`
}

func checkUserName(timeout time.Duration, service AuthService, slog *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := checkUserNameRequest{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if !strings.Contains(err.Error(), "EOF") {
				slog.Debug("Can't decode body from sign-up request", "err", err.Error())

				err := responseJSONError(w, http.StatusBadRequest, 0, "", "Bad request")
				if err != nil {
					slog.Error("Can't make response", "err", err.Error())
				}
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		isExists, err := service.CheckUserName(ctx, req.UserName)
		if err != nil {
			slog.Error("Can't check user name", "err", err.Error())

			err = responseJSONError(w, http.StatusInternalServerError, 0, "", "Internal error")
			if err != nil {
				slog.Error("Can't make response", "err", err.Error())
			}
			return
		}

		err = responseCheckJSON(w, http.StatusOK, isExists)
		if err != nil {
			slog.Error("Can't make response", "err", err.Error())
		}
	}
}
