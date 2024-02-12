package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func GetAccessToken(oauth2Cfg clientcredentials.Config, tokenType ...string) (string, error) {
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

	if len(tokenType) > 0 {
		if len(oauth2Cfg.EndpointParams) == 0 {
			oauth2Cfg.EndpointParams = make(url.Values)
		}
		oauth2Cfg.EndpointParams.Add("token_format", tokenType[0])
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	token, err := oauth2Cfg.Token(ctx)
	if err != nil {
		return "", err
	}
	if !token.Valid() {
		return "", fmt.Errorf("token invalid. got: %#v", token)
	}
	if token.TokenType != "Bearer" {
		return "", fmt.Errorf("token type = %q; want %q", token.TokenType, "Bearer")
	}
	return token.AccessToken, nil
}
