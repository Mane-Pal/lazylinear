package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	v, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 32 bytes → 43 base64url characters (no padding)
	if len(v) != 43 {
		t.Errorf("expected length 43, got %d", len(v))
	}

	// Must be base64url-safe (no +, /, or =)
	if strings.ContainsAny(v, "+/=") {
		t.Errorf("verifier contains non-base64url characters: %s", v)
	}
}

func TestGenerateCodeVerifierUniqueness(t *testing.T) {
	v1, _ := GenerateCodeVerifier()
	v2, _ := GenerateCodeVerifier()
	if v1 == v2 {
		t.Error("two verifiers should not be identical")
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

	// Compute expected challenge manually
	h := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(h[:])

	got := GenerateCodeChallenge(verifier)
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestGenerateCodeChallengeBase64URLSafe(t *testing.T) {
	verifier, _ := GenerateCodeVerifier()
	challenge := GenerateCodeChallenge(verifier)

	if strings.ContainsAny(challenge, "+/=") {
		t.Errorf("challenge contains non-base64url characters: %s", challenge)
	}
}
