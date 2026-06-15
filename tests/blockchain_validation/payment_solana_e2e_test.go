package integration

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/daniel0321forever/terriyaki-go/internal/application/dto"
	"github.com/daniel0321forever/terriyaki-go/internal/application/services"
	"github.com/daniel0321forever/terriyaki-go/internal/domain/entities"
	"github.com/daniel0321forever/terriyaki-go/internal/infrastructure/db/postgres"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := resp.Body.Close(); err != nil {
		t.Logf("warning: failed to close response body: %v", err)
	}
	return resp, bodyBytes
}

// TestSolanaE2E runs a minimal happy-path integration flow against a running backend
// and a local Solana validator.
// Requirements:
// - backend running at http://localhost:8080
// - `solana` CLI available in PATH
// - SOLANA_ORACLE_PRIVATE_KEY is set to the local signing key for this test
func TestSolanaE2E(t *testing.T) {
	// reuse parsing logic from services package to avoid duplication
	rpcEndpoint, programID, oraclePubkey, oraclePrivateKey, err := services.LoadSolanaConfigFromEnv()
	if err != nil {
		t.Skipf("skipping solana E2E: %v", err)
	}
	base := "http://localhost:8080"

	// derive wallet pubkey from the shared local signing key
	wallet := oraclePrivateKey.PublicKey().String()
	t.Logf("using wallet %s", wallet)

	// fund wallet on localnet
	airdropCmd := exec.Command("solana", "airdrop", "2", wallet, "--url", rpcEndpoint)
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

	// 2) add solana wallet payment method (use programID from env helper)
	addBody := map[string]interface{}{
		"method_type":    "solana_wallet",
		"wallet_address": wallet,
		"network":        "localnet",
		"program_id":     programID.String(),
	}
	headers := map[string]string{"Authorization": "Bearer " + token}
	resp, body = httpPostJSON(t, base+"/api/v1/payments/methods", addBody, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add payment method failed status=%d body=%s", resp.StatusCode, string(body))
	}

	// 3) create a collection intent (unsigned tx)
	pledgeID := fmt.Sprintf("pledge-%d", time.Now().UnixNano())
	amountLamports := int64(1000000) // 0.001 SOL on localnet
	deadline := time.Now().Add(30 * time.Minute).Unix()

	intentBody := map[string]interface{}{
		"wallet_address":  wallet,
		"network":         "localnet",
		"program_id":      programID.String(),
		"pledge_id":       pledgeID,
		"oracle_pubkey":   oraclePubkey.String(),
		"amount_lamports": amountLamports,
		"deadline_unix":   deadline,
	}
	if idv, ok := reg["id"].(string); ok && idv != "" {
		intentBody["user_id"] = idv
	} else {
		intentBody["user_id"] = "test-user"
	}
	idemp := fmt.Sprintf("e2e-collection-%d", time.Now().UnixNano())
	headers["Idempotency-Key"] = idemp
	resp, body = httpPostJSON(t, base+"/api/v1/payments/solana/collection-intent", intentBody, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("collection intent failed status=%d body=%s", resp.StatusCode, string(body))
	}
	var intentResp map[string]any
	if err := json.Unmarshal(body, &intentResp); err != nil {
		t.Fatalf("unmarshal intent response: %v", err)
	}
	providerRef, _ := intentResp["provider_reference"].(string)
	pledgePDA, _ := intentResp["pledge_pda"].(string)

	// 4) use the backend-provided unsigned_tx_json to build the transaction
	unsignedJSONStr, _ := intentResp["unsigned_tx_json"].(string)
	if unsignedJSONStr == "" {
		t.Fatalf("missing unsigned_tx_json in intent response: %v", intentResp)
	}

	var unsignedEnv struct {
		ProgramID       string   `json:"program_id"`
		Accounts        []string `json:"accounts"`
		InstructionB64  string   `json:"instruction_b64"`
		RecentBlockhash string   `json:"recent_blockhash"`
	}
	if err := json.Unmarshal([]byte(unsignedJSONStr), &unsignedEnv); err != nil {
		t.Fatalf("parse unsigned_tx_json: %v", err)
	}

	instrData, err := base64.StdEncoding.DecodeString(unsignedEnv.InstructionB64)
	if err != nil {
		t.Fatalf("decode instruction_b64: %v", err)
	}

	acctSlice := solana.AccountMetaSlice{}
	for _, a := range unsignedEnv.Accounts {
		pk, err := solana.PublicKeyFromBase58(a)
		if err != nil {
			t.Fatalf("invalid account in unsigned_tx_json: %v", err)
		}
		isSigner := a == wallet
		isWritable := a == pledgePDA
		acctSlice = append(acctSlice, &solana.AccountMeta{PublicKey: pk, IsSigner: isSigner, IsWritable: isWritable})
	}

	progPub, err := solana.PublicKeyFromBase58(unsignedEnv.ProgramID)
	if err != nil {
		t.Fatalf("invalid program id in unsigned_tx_json: %v", err)
	}

	rpcClient := rpc.New(rpcEndpoint)
	bhRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		t.Fatalf("fetch recent blockhash: %v", err)
	}
	recent := bhRes.Value.Blockhash

	tx, err := solana.NewTransaction(
		[]solana.Instruction{solana.NewInstruction(progPub, acctSlice, instrData)},
		recent,
	)
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	priv := ed25519.PrivateKey(oraclePrivateKey[:])
	msg, err := tx.Message.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal tx message: %v", err)
	}
	sig := ed25519.Sign(priv, msg)
	tx.Signatures = []solana.Signature{solana.SignatureFromBytes(sig)}
	signedBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal signed tx: %v", err)
	}
	signedBase64 := base64.StdEncoding.EncodeToString(signedBytes)

	// 5) submit signed tx to backend for broadcast
	submitBody := map[string]interface{}{
		"provider_reference":        providerRef,
		"signed_transaction_base64": signedBase64,
		"network":                   "localnet",
	}
	headers["Idempotency-Key"] = idemp
	resp, body = httpPostJSON(t, base+"/api/v1/payments/solana/submit-signed-transaction", submitBody, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("submit signed tx failed status=%d body=%s", resp.StatusCode, string(body))
	}

	var submitResp map[string]any
	if err := json.Unmarshal(body, &submitResp); err != nil {
		t.Fatalf("unmarshal submit response: %v", err)
	}
	sigStr, _ := submitResp["signature"].(string)
	if sigStr == "" {
		t.Fatalf("no signature returned from submit response: %s", string(body))
	}

	// poll for finalized signature
	sigObj, err := solana.SignatureFromBase58(sigStr)
	if err != nil {
		t.Fatalf("invalid signature format: %v", err)
	}
	client := rpc.New(rpcEndpoint)
	timeout := time.After(60 * time.Second)
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	finalized := false
	for {
		select {
		case <-timeout:
			t.Fatalf("timeout waiting for signature finalization %s", sigStr)
		case <-tick.C:
			statuses, err := client.GetSignatureStatuses(context.Background(), true, sigObj)
			if err != nil {
				continue
			}
			if len(statuses.Value) > 0 && statuses.Value[0] != nil && statuses.Value[0].ConfirmationStatus == rpc.ConfirmationStatusFinalized {
				finalized = true
			}
		}
		if finalized {
			break
		}
	}

	// check pledge account exists and has lamports (initialized)
	pledgePubKey, _ := solana.PublicKeyFromBase58(pledgePDA)
	acctBefore, err := client.GetAccountInfo(context.Background(), pledgePubKey)
	if err != nil {
		t.Fatalf("get pledge account info before resolution: %v", err)
	}
	if acctBefore.Value == nil || acctBefore.Value.Lamports == 0 {
		t.Fatalf("expected pledge account to be created and funded, got nil or zero lamports")
	}
	t.Logf("pledge account initialized with %d lamports", acctBefore.Value.Lamports)

	// 5.5) Resolve the pledge as oracle
	userRepo := postgres.NewGormUserRepository(postgres.Db)
	grindRepo := postgres.NewGormGrindRepository(postgres.Db)
	participationRepo := postgres.NewGormParticipationRepository(postgres.Db)
	paymentInfoRepo := postgres.NewGormStripePaymentInfoRepository(postgres.Db)
	paymentIdempotencyRepo := postgres.NewGormPaymentIdempotencyRepository(postgres.Db)
	paymentSettlementRepo := postgres.NewGormPaymentSettlementRepository(postgres.Db)

	paymentFactory := services.NewPaymentServiceFactory(
		userRepo,
		grindRepo,
		participationRepo,
		paymentInfoRepo,
		paymentIdempotencyRepo,
		paymentSettlementRepo,
	)
	solanaPaymentService, err := paymentFactory.BuildForProvider(entities.PaymentProviderSolana)
	if err != nil {
		t.Fatalf("build Solana payment service: %v", err)
	}

	resolveIdemp := fmt.Sprintf("e2e-resolve-%d", time.Now().UnixNano())
	resolveDTO, err := dto.NewSolanaResolvePledgeDTO(
		"solana_resolve_pledge",
		"success",
		"", // penaltyPoolKey (not needed for success)
		pledgePDA,
		wallet,
		"localnet",
		"proof-hash-abc-123",
	)
	if err != nil {
		t.Fatalf("build resolve DTO: %v", err)
	}

	// Create the settlement record in DB first
	dbUserID, _ := reg["id"].(string)
	settlement := entities.NewPaymentSettlement(
		dbUserID,
		"solana_resolve_pledge",
		resolveIdemp,
		entities.PaymentProviderSolana,
		wallet,
		amountLamports,
	)
	settlement.Status = entities.SettlementStatusPending
	_, err = paymentSettlementRepo.Create(settlement)
	if err != nil {
		t.Fatalf("failed to create settlement record: %v", err)
		return
	}

	resolveRes, err := solanaPaymentService.ResolvePledgeAsOracle(resolveDTO, resolveIdemp)
	if err != nil {
		t.Fatalf("ResolvePledgeAsOracle failed: %v", err)
		return
	}
	t.Logf("Oracle resolve success signature: %s", resolveRes.Signature)

	// Verify that the on-chain balance decreased by the escrow amount (1,000,000 lamports)
	acctAfter, err := client.GetAccountInfo(context.Background(), pledgePubKey)
	if err != nil {
		t.Fatalf("get pledge account info after resolution: %v", err)
		return
	}
	if acctAfter.Value == nil {
		t.Fatalf("expected pledge account to exist after resolution")
		return
	}
	diff := int64(acctBefore.Value.Lamports) - int64(acctAfter.Value.Lamports)
	if diff != amountLamports {
		t.Fatalf("expected on-chain balance to decrease by %d, decreased by %d (before=%d, after=%d)", amountLamports, diff, acctBefore.Value.Lamports, acctAfter.Value.Lamports)
		return
	}
	t.Logf("verified on-chain balance decreased by %d to %d", diff, acctAfter.Value.Lamports)

	// Verify the database record has transitioned to settled_onchain
	settlement, err = paymentSettlementRepo.FindByOperationAndKey("solana_resolve_pledge", resolveIdemp)
	if err != nil {
		t.Fatalf("failed to query settlement from DB: %v", err)
		return
	}
	if settlement == nil {
		t.Fatalf("settlement record not found in DB")
		return
	}
	if settlement.Status != entities.SettlementStatusSettledOnChain {
		t.Fatalf("expected settlement status to be %s, got %s", entities.SettlementStatusSettledOnChain, settlement.Status)
		return
	}
	t.Logf("verified database settlement record status is %s", settlement.Status)
}
