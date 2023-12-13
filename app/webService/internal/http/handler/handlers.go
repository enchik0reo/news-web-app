package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"newsWebApp/app/webService/internal/services"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

var (
	ErrInvalidCredetials = "invalid credentials"
)

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

		id, err := service.SaveUser(req.Email, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrUserExists) {
				http.Error(w, "user already exists", http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
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

		id, acsToken, refToken, err := service.LoginUser(req.Email, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrUserDoesntExists) {
				http.Error(w, ErrInvalidCredetials, http.StatusUnauthorized)
				return
			}
			http.Error(w, "failed to get user", http.StatusInternalServerError)
			return
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
