package utils

import "net/http"

func GetNextUrl(r *http.Request, defaultUrl string) string {
	nextUrl := r.Referer()

	if nextUrl == "" {
		nextUrl = defaultUrl
	}

	return nextUrl
}
