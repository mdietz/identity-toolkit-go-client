// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitkit

import (
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
)

// ServiceAccountTransport is an implementation of http.RoundTripper that can
// automatically fetch an access token for the service account to access
// identitytoolkit service.
type ServiceAccountTransport struct {
	// Assertion is the JWT assertion generated by the service account.
	Assertion *jwt.Token
	// Token is the identitytookit API OAuth2 token.
	Token *oauth.Token
	// Transport is the underlying HTTP transport.
	Transport http.RoundTripper
}

// transport returns the underlying HTTP transport. If it is not specified, an
// http.DefaultTransport is used.
func (t *ServiceAccountTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// refreshToken fetches the OAuth2 access token for the service account if there
// is no access token or it is expired.
func (t *ServiceAccountTransport) refreshToken() error {
	if t.Token == nil || t.Token.Expired() {
		token, err := t.Assertion.Assert(&http.Client{Transport: t.transport()})
		if err != nil {
			return err
		}
		t.Token = token
	}
	return nil
}

// RoundTrip implements the http.RoundTripper interface.
func (t *ServiceAccountTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Refresh the access token if necessary.
	if err := t.refreshToken(); err != nil {
		return nil, err
	}
	// Copy the request to avoid modifying the original request.
	// This is required by the specification of http.RoundTripper.
	newReq := *req
	newReq.Header = make(http.Header)
	for k, v := range req.Header {
		newReq.Header[k] = v
	}
	// Add Additional headers.
	newReq.Header.Set("Authorization", "Bearer "+t.Token.AccessToken)
	newReq.Header.Set("User-Agent", "gitkit-go-client/0.1")
	newReq.Header.Set("Content-Type", "application/json")
	return t.transport().RoundTrip(&newReq)
}
