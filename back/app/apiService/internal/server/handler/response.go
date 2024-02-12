package handler

import (
	"encoding/json"
	"fmt"
	"newsWebApp/app/apiService/internal/models"
)

type response struct {
	Status  int    `json:"status"`
	AcToken string `json:"access_token,omitempty"`
	Body    body   `json:"body"`
}

type body struct {
	UserID   string           `json:"uid,omitempty"`
	Articles []models.Article `json:"articles,omitempty"`
	Error    string           `json:"error,omitempty"`
}

func makeResponse(status int, uid string, acToken string, articles []models.Article, error string) ([]byte, error) {
	resp := response{}

	resp.Status = status

	if acToken != "" {
		resp.AcToken = acToken
	}

	if uid != "" {
		resp.Body.UserID = uid
	}

	if len(articles) > 0 && error != "" {
		return nil, fmt.Errorf("can't pack articles and error at the same time")
	}

	if len(articles) > 0 {
		resp.Body.Articles = articles
	}

	if error != "" {
		resp.Body.Error = error
	}

	return json.Marshal(resp)
}
