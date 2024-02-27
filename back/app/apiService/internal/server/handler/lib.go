package handler

import "net/http"

type (
	userID string
	userNm string
	token  string
)

var (
	uid         userID = "uid"
	userName    userNm = "user_name"
	accessToken token  = "access_token"
)

func getInfoFromCtx(r *http.Request) (int64, string, string) {
	uID, ok := r.Context().Value(uid).(int64)
	if !ok {
		uID = 0
	}

	acsToken, ok := r.Context().Value(accessToken).(string)
	if !ok {
		acsToken = ""
	}

	uName, ok := r.Context().Value(userName).(string)
	if !ok {
		uName = ""
	}

	return uID, uName, acsToken
}
