package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"newsWebApp/app/apiService/internal/models"
)

type response struct {
	Status int      `json:"status"`
	Body   respBody `json:"body"`
}

type respBody struct {
	UserID   int64            `json:"uid,omitempty"`
	AcToken  string           `json:"access_token,omitempty"`
	Articles []models.Article `json:"articles,omitempty"`
	Error    string           `json:"error,omitempty"`
}

func responseJSON(w http.ResponseWriter, status int, uID int64, acsToken string, articles []models.Article) error {
	resp := response{
		Status: status,
	}

	if acsToken != "" {
		resp.Body.AcToken = acsToken
	}

	if uID != 0 {
		resp.Body.UserID = uID
	}

	if len(articles) > 0 {
		resp.Body.Articles = articles
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

func socketResponse(w io.WriteCloser, status int, articles []models.Article) error {
	resp := response{
		Status: status,
	}

	if len(articles) > 0 {
		resp.Body.Articles = articles
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = w.Write(respJSON)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}
