package main

import (
	"bytes"
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
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Firebase Cloud Messaging (HTTP v1) — yeni məhsul əlavə olunanda "new_products"
// topic-inə push göndərir. Yalnız stdlib (service account JSON → JWT → access token → send).

type fcmServiceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
	ProjectID   string `json:"project_id"`
}

var (
	fcmMu       sync.Mutex
	fcmSA       *fcmServiceAccount
	fcmLoaded   bool
	fcmToken    string
	fcmTokenExp time.Time
)

// FCM_SERVICE_ACCOUNT env-i (raw JSON və ya base64) → service account
func loadFcmSA() *fcmServiceAccount {
	fcmMu.Lock()
	defer fcmMu.Unlock()
	if fcmLoaded {
		return fcmSA
	}
	fcmLoaded = true
	raw := strings.TrimSpace(os.Getenv("FCM_SERVICE_ACCOUNT"))
	if raw == "" {
		return nil
	}
	data := []byte(raw)
	if !strings.HasPrefix(raw, "{") {
		if dec, err := base64.StdEncoding.DecodeString(raw); err == nil {
			data = dec
		}
	}
	var sa fcmServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil || sa.ClientEmail == "" || sa.PrivateKey == "" {
		log.Printf("FCM: service account oxunmadı (FCM_SERVICE_ACCOUNT)")
		return nil
	}
	if sa.TokenURI == "" {
		sa.TokenURI = "https://oauth2.googleapis.com/token"
	}
	fcmSA = &sa
	log.Printf("FCM: hazır (project %s)", sa.ProjectID)
	return fcmSA
}

func fcmAccessToken(sa *fcmServiceAccount) (string, error) {
	fcmMu.Lock()
	defer fcmMu.Unlock()
	if fcmToken != "" && time.Now().Before(fcmTokenExp) {
		return fcmToken, nil
	}
	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("private key pem")
	}
	var key *rsa.PrivateKey
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rk, ok := k.(*rsa.PrivateKey); ok {
			key = rk
		}
	}
	if key == nil {
		if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			key = k
		}
	}
	if key == nil {
		return "", fmt.Errorf("private key parse")
	}
	now := time.Now()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(
		`{"iss":"%s","scope":"https://www.googleapis.com/auth/firebase.messaging","aud":"%s","iat":%d,"exp":%d}`,
		sa.ClientEmail, sa.TokenURI, now.Unix(), now.Add(time.Hour).Unix())))
	signingInput := header + "." + claims
	h := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	jwt := signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
	resp, err := http.PostForm(sa.TokenURI, url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwt},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token status %d: %s", resp.StatusCode, string(body))
	}
	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	json.Unmarshal(body, &tr)
	if tr.AccessToken == "" {
		return "", fmt.Errorf("boş access token")
	}
	fcmToken = tr.AccessToken
	fcmTokenExp = now.Add(time.Duration(tr.ExpiresIn-60) * time.Second)
	return fcmToken, nil
}

// sendNewProductPush — "new_products" topic-inə push. FCM qurulmayıbsa səssiz keçir.
func sendNewProductPush(title, body string) {
	sa := loadFcmSA()
	if sa == nil {
		return
	}
	token, err := fcmAccessToken(sa)
	if err != nil {
		log.Printf("FCM token: %v", err)
		return
	}
	msg := map[string]any{
		"message": map[string]any{
			"topic": "new_products",
			"notification": map[string]any{
				"title": title,
				"body":  body,
			},
			"android": map[string]any{"priority": "high"},
		},
	}
	b, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", "https://fcm.googleapis.com/v1/projects/"+sa.ProjectID+"/messages:send", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		log.Printf("FCM send: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		log.Printf("FCM send status %d: %s", resp.StatusCode, string(rb))
	}
}
