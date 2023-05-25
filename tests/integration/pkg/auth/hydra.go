package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func GetHydraOAuth2AccessToken(oauth2Cfg clientcredentials.Config) (string, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return "", err
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 10,
		Jar:     jar,
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	token, err := oauth2Cfg.Token(ctx)
	if err != nil {
		return "", err
	}
	if !token.Valid() {
		return "", fmt.Errorf("token invalid. got: %#v", token)
	}
	if token.TokenType != "bearer" {
		return "", fmt.Errorf("token type = %q; want %q", token.TokenType, "bearer")
	}
	return token.AccessToken, nil
}
