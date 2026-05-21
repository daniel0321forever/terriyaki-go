package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func httpPostJSON(t *testing.T, url string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, bodyBytes
}

// TestSolanaE2E runs a minimal happy-path integration flow against a running backend
// and a local Solana validator.
// Requirements:
// - backend running at http://localhost:8080
// - `solana-keygen` and `solana` CLI available in PATH
// - local keypair at ~/.config/solana/id.json (or pass SOLANA_KEYPAIR env)
func TestSolanaE2E(t *testing.T) {
	base := "http://localhost:8080"

	rpc := os.Getenv("SOLANA_RPC_ENDPOINT")
	if rpc == "" {
		t.Fatal("SOLANA_RPC_ENDPOINT env var is required, e.g. http://127.0.0.1:8899")
	}

	// locate keypair
	kp := os.ExpandEnv("$HOME/.config/solana/id.json")
	if _, err := os.Stat(kp); err != nil {
		t.Skipf("solana keypair not found at %s: %v", kp, err)
	}

	// derive wallet pubkey
	out, err := exec.Command("solana-keygen", "pubkey", kp).CombinedOutput()
	if err != nil {
		t.Skipf("solana-keygen not available or failed: %v output=%s", err, string(out))
	}
	wallet := strings.TrimSpace(string(out))
	t.Logf("using wallet %s", wallet)

	// fund wallet on localnet
	airdropCmd := exec.Command("solana", "airdrop", "2", wallet, "--url", rpc)
	if out, err := airdropCmd.CombinedOutput(); err != nil {
		t.Logf("airdrop warning: %v output=%s", err, string(out))
	}

	// 1) register user
	username := fmt.Sprintf("e2e-%d", time.Now().UnixNano())
	regBody := map[string]string{"username": username, "email": username + "@example.com", "password": "password"}
	resp, body := httpPostJSON(t, base+"/api/v1/register", regBody, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register returned status %d body=%s", resp.StatusCode, string(body))
	}
	var reg map[string]interface{}
	if err := json.Unmarshal(body, &reg); err != nil {
		t.Fatalf("unmarshal register response: %v", err)
	}
	token, _ := reg["token"].(string)
	if token == "" {
		t.Fatalf("no token in register response: %s", string(body))
	}

	// 2) add solana wallet payment method
	addBody := map[string]interface{}{
		"method_type":    "solana_wallet",
		"wallet_address": wallet,
		"network":        "localnet",
		"program_id":     "BgNjXioQqVNNihH4QCtjthDKAynZLVDixArQgmY7oRM4",
	}
	headers := map[string]string{"Authorization": "Bearer " + token}
	resp, body = httpPostJSON(t, base+"/api/v1/payments/methods", addBody, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add payment method failed status=%d body=%s", resp.StatusCode, string(body))
	}

	// Success: the backend accepted the wallet method. Further settlement flows
	// (create intent, sign, submit) are exercised in other targeted tests.
}
