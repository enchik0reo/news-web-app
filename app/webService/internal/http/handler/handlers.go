package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"newsWebApp/app/webService/internal/services"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It's a home page"))
	}
}

func signUp(service AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(Request)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		id, err := service.SaveUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserExists):
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case errors.Is(err, services.ErrInvalidValue):
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	}
}

func signIn(service AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(Request)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		id, acsToken, refToken, err := service.LoginUser(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserDoesntExists):
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			case errors.Is(err, services.ErrInvalidValue):
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			default:
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Authorization", "Bearer "+acsToken)

		ck := http.Cookie{
			Name:     "refresh_token",
			Domain:   r.URL.Host,
			Path:     "/",
			Value:    refToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}

		http.SetCookie(w, &ck)

		w.Header().Add("id", fmt.Sprint(id))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"Access Token":  "Bearer " + acsToken,
			"Refresh Token": refToken,
		})
	}
}

func home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := w.Header().Get("id")
		w.Write([]byte("You are user with id = " + id))
	}
}
