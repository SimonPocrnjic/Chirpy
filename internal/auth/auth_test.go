package auth

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	hashRX, err := regexp.Compile(`^\$argon2id\$v=19\$m=65536,t=1,p=[0-9]{1,4}\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`)
	if err != nil {
		t.Fatal(err)
	}

	hash1, err := HashPassword("pa$$word")
	if err != nil {
		t.Fatal(err)
	}

	if !hashRX.MatchString(hash1) {
		t.Errorf("hash %q not in correct format", hash1)
	}

	hash2, err := HashPassword("pa$$word")

	if strings.Compare(hash1, hash2) == 0 {
		t.Error("hashes must be unique")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("pa$$word")
	if err != nil {
		t.Fatal(err)
	}

	match, err := CheckPasswordHash("pa$$word", hash)
	if err != nil {
		t.Fatal(err)
	}

	if !match {
		t.Error("expected password and hash to match")
	}

	match, err = CheckPasswordHash("otherPa$$word", hash)
	if err != nil {
		t.Fatal(err)
	}

	if match {
		t.Error("expected password and hash to not match")
	}
}

func TestMakeJWT(t *testing.T) {
	uuid := uuid.MustParse("32d1ae43-502f-4bbc-99c5-0076be6921ef")
	tokenSecret := "B7FfEklt1igu77YlwbGg3FLtuxYfO9Tuemf7ZZ2LasMVBAuEA9RuHUeVg1qIKhZA"
	expiresIn, _ := time.ParseDuration("6h")

	token, err := MakeJWT(uuid, tokenSecret, expiresIn)

	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Error("did not expect empty token")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := &http.Header{}

	headers.Set("Authorization", "Bearer wz+znvOhhL2OVc2/HwD2qHR6iDORsIykUwYs8eD3uFRk8BKCWC1T+vaCWjuD26bCj1Oag3odAwYQpmBLclaysA==")

	token, err := GetBearerToken(*headers)

	if err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("Bearer token: %s", token)
	}
}
