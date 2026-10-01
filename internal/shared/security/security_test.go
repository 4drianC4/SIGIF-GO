package security

import "testing"

func TestHashPasswordCanBeVerified(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if err := VerifyPassword(password, hash); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if err := VerifyPassword("wrong password", hash); err == nil {
		t.Fatal("VerifyPassword() accepted an incorrect password")
	}
}
