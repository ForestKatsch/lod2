package utils

import (
	"net/http"
	"net/url"
)

func GetNextUrl(r *http.Request, defaultUrl string) string {
	nextUrl := r.Referer()

	if nextUrl == "" {
		nextUrl = defaultUrl
	}

	return nextUrl
}

func UrlDecode(str string) (string, error) {
	decoded, err := url.QueryUnescape(str)
	if err != nil {
		return "", err
	}
	return decoded, nil
}
