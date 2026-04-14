package githubapp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMintInstallationTokenUsesJWTAndRepoScope(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	var (
		seenAuth  string
		seenRepos []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app/installations/456/access_tokens" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		seenAuth = r.Header.Get("Authorization")
		var body map[string][]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		seenRepos = body["repositories"]
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":      "installation-token",
			"expires_at": "2026-04-14T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient("123", "456", string(pemBytes), server.URL, server.Client())
	client.now = func() time.Time { return time.Unix(1_700_000_000, 0).UTC() }

	token, err := client.MintInstallationToken(context.Background(), []string{"rsi-agent-platform"})
	if err != nil {
		t.Fatalf("MintInstallationToken() error = %v", err)
	}
	if token.Token != "installation-token" {
		t.Fatalf("unexpected token %#v", token)
	}
	if len(seenRepos) != 1 || seenRepos[0] != "rsi-agent-platform" {
		t.Fatalf("unexpected repository scope %#v", seenRepos)
	}
	if !strings.HasPrefix(seenAuth, "Bearer ") {
		t.Fatalf("unexpected auth header %q", seenAuth)
	}
	if len(strings.Split(strings.TrimPrefix(seenAuth, "Bearer "), ".")) != 3 {
		t.Fatalf("expected JWT auth header, got %q", seenAuth)
	}
}

func TestMintInstallationTokenSupportsEscapedNewlines(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":      "installation-token",
			"expires_at": "2026-04-14T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient("123", "456", strings.ReplaceAll(string(pemBytes), "\n", `\n`), server.URL, server.Client())
	if _, err := client.MintInstallationToken(context.Background(), nil); err != nil {
		t.Fatalf("MintInstallationToken() error = %v", err)
	}
}
