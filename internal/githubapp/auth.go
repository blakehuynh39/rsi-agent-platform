package githubapp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	appID          string
	installationID string
	privateKeyPEM  string
	apiBaseURL     string
	httpClient     *http.Client
	now            func() time.Time
}

type InstallationToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type installationTokenRequest struct {
	Repositories []string `json:"repositories,omitempty"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type jwtClaims struct {
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
	Issuer    int64 `json:"iss"`
}

func NewClient(appID string, installationID string, privateKeyPEM string, apiBaseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		appID:          strings.TrimSpace(appID),
		installationID: strings.TrimSpace(installationID),
		privateKeyPEM:  strings.TrimSpace(privateKeyPEM),
		apiBaseURL:     strings.TrimRight(strings.TrimSpace(apiBaseURL), "/"),
		httpClient:     httpClient,
		now:            time.Now,
	}
}

func (c *Client) MintInstallationToken(ctx context.Context, repositories []string) (InstallationToken, error) {
	if c.appID == "" {
		return InstallationToken{}, fmt.Errorf("github app id is required")
	}
	if c.installationID == "" {
		return InstallationToken{}, fmt.Errorf("github app installation id is required")
	}
	if c.privateKeyPEM == "" {
		return InstallationToken{}, fmt.Errorf("github app private key is required")
	}
	if c.apiBaseURL == "" {
		return InstallationToken{}, fmt.Errorf("github api base url is required")
	}

	jwt, err := c.signedJWT()
	if err != nil {
		return InstallationToken{}, err
	}

	body := installationTokenRequest{}
	filteredRepos := make([]string, 0, len(repositories))
	for _, repo := range repositories {
		repo = strings.TrimSpace(repo)
		if repo != "" {
			filteredRepos = append(filteredRepos, repo)
		}
	}
	if len(filteredRepos) > 0 {
		body.Repositories = filteredRepos
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return InstallationToken{}, fmt.Errorf("marshal github app access token request: %w", err)
	}
	endpoint := fmt.Sprintf("%s/app/installations/%s/access_tokens", c.apiBaseURL, c.installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return InstallationToken{}, fmt.Errorf("build github app access token request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return InstallationToken{}, fmt.Errorf("request github app access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return InstallationToken{}, fmt.Errorf("github app access token request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var token InstallationToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return InstallationToken{}, fmt.Errorf("decode github app access token response: %w", err)
	}
	if strings.TrimSpace(token.Token) == "" {
		return InstallationToken{}, fmt.Errorf("github app access token response missing token")
	}
	return token, nil
}

func (c *Client) signedJWT() (string, error) {
	privateKey, err := parsePrivateKey(c.privateKeyPEM)
	if err != nil {
		return "", err
	}
	now := c.now().UTC()
	header := jwtHeader{
		Algorithm: "RS256",
		Type:      "JWT",
	}
	issuer, err := issuerClaim(c.appID)
	if err != nil {
		return "", err
	}
	claims := jwtClaims{
		IssuedAt:  now.Add(-60 * time.Second).Unix(),
		ExpiresAt: now.Add(9 * time.Minute).Unix(),
		Issuer:    issuer,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal github app jwt header: %w", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal github app jwt claims: %w", err)
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign github app jwt: %w", err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func parsePrivateKey(value string) (*rsa.PrivateKey, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), `\n`, "\n")
	block, _ := pem.Decode([]byte(normalized))
	if block == nil {
		return nil, fmt.Errorf("github app private key PEM decode failed")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("github app private key parse failed: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("github app private key must be RSA")
	}
	return key, nil
}

func issuerClaim(appID string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(appID), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("github app id must be numeric: %q", strings.TrimSpace(appID))
	}
	return id, nil
}
