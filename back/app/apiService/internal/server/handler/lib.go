package handler

import "net/http"

type (
	userID string
	token  string
)

var (
	uid         userID = "uid"
	accessToken token  = "access_token"
)

func getInfoFromCtx(r *http.Request) (int64, string) {
	uID, ok := r.Context().Value(uid).(int64)
	if !ok {
		uID = 0
	}

	acsToken, ok := r.Context().Value(accessToken).(string)
	if !ok {
		acsToken = ""
	}

	return uID, acsToken
}
