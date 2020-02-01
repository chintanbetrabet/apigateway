package auth

import (
	"fmt"
	"net/http"
	"redis_local"
)

func tokenAuthMethod() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		return_text := ""
		token := r.Header.Get("X-Auth-Token")
		if token != "" {
			return_text += "Token is not null:" + token + "\n"
			search_key := fmt.Sprintf("auth_token:%s", token)
			account_id := redis_local.GetStringKeyFromMap(search_key, "account_id")
			if account_id == "" {
				return_text += "Could not validate account using the given token" + "\n"
			} else {
				return_text += "Authenicated, user is from Account id:" + account_id + "\n"
				fmt.Fprintln(w, return_text)
				fmt.Printf("\n%+v\n", w)
				return
			}
		} else {
			return_text = "Token is empty"
		}
		http.Error(w, return_text, http.StatusForbidden)
		fmt.Println("%+v", w)
	}
}
