package handler

import (
	"encoding/json"
	"net/http"

	"newsWebApp/app/apiService/internal/models"
)

type response struct {
	Status int      `json:"status"`
	Body   respBody `json:"body"`
}

type respBody struct {
	UserID   int64            `json:"uid,omitempty"`
	UserName string           `json:"user_name,omitempty"`
	AcToken  string           `json:"access_token,omitempty"`
	Articles []models.Article `json:"articles,omitempty"`
	Error    string           `json:"error,omitempty"`
	Exists   bool             `json:"exists,omitempty"`
}

func responseJSONOk(w http.ResponseWriter, status int, body respBody) error {
	resp := response{
		Status: status,
		Body:   body,
	}

	w.Header().Add("Content-Type", "application/json")

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = w.Write(respJSON)
	if err != nil {
		return err
	}

	return nil
}

func responseCheckJSON(w http.ResponseWriter, status int, isExists bool) error {
	resp := response{
		Status: status,
	}

	resp.Body.Exists = isExists

	w.Header().Add("Content-Type", "application/json")

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = w.Write(respJSON)
	if err != nil {
		return err
	}

	return nil
}

func responseJSONError(w http.ResponseWriter, status int, uID int64, acsToken string, error string) error {
	resp := response{
		Status: status,
	}

	if acsToken != "" {
		resp.Body.AcToken = acsToken
	}

	if uID != 0 {
		resp.Body.UserID = uID
	}

	if error != "" {
		resp.Body.Error = error
	}

	w.Header().Add("Content-Type", "application/json")

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = w.Write(respJSON)
	if err != nil {
		return err
	}

	return nil
}
