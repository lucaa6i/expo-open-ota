package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
)

func TestCreateHash(t *testing.T) {
	data := []byte("test data")
	tests := []struct {
		name      string
		algorithm string
		encoding  string
		expectErr bool
	}{
		{"SHA256 Hex", "sha256", "hex", false},
		{"SHA512 Base64", "sha512", "base64", false},
		{"MD5 Hex", "md5", "hex", false},
		{"Unsupported Algorithm", "sha1", "hex", true},
		{"Unsupported Encoding", "sha256", "binary", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := CreateHash(data, tt.algorithm, tt.encoding)
			if (err != nil) != tt.expectErr {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && hash == "" {
				t.Errorf("expected non-empty hash")
			}
		})
	}
}

func TestConvertSHA256HashToUUID(t *testing.T) {
	input := "1234567890abcdef1234567890abcdef12345678"
	expected := "12345678-90ab-cdef-1234-567890abcdef"
	result := ConvertSHA256HashToUUID(input)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}

	shortInput := "short"
	result = ConvertSHA256HashToUUID(shortInput)
	if result != "" {
		t.Errorf("expected empty string for short input, got %s", result)
	}
}

func TestGetBase64URLEncoding(t *testing.T) {
	input := base64.StdEncoding.EncodeToString([]byte("test data"))
	expected := strings.ReplaceAll(strings.ReplaceAll(input, "+", "-"), "/", "_")
	expected = strings.TrimRight(expected, "=")
	result := GetBase64URLEncoding(input)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestSignRSASHA256(t *testing.T) {
	data := "test data"

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	signature, err := SignRSASHA256(data, string(privateKeyPEM))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		t.Fatalf("failed to decode signature: %v", err)
	}

	hash := sha256.Sum256([]byte(data))
	err = rsa.VerifyPKCS1v15(&privateKey.PublicKey, crypto.SHA256, hash[:], signatureBytes)
	if err != nil {
		t.Errorf("signature verification failed: %v", err)
	}

	invalidPEM := "-----BEGIN PRIVATE KEY-----\ninvalid\n-----END PRIVATE KEY-----"
	_, err = SignRSASHA256(data, invalidPEM)
	if err == nil {
		t.Errorf("expected error for invalid private key, got none")
	}
}
