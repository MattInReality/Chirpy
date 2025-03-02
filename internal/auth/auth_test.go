package auth

import (
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

const testPassword = "How do you do"
const tokenSecret = "my secret"

var testHash string
var testToken string

var testUUID uuid.UUID

func TestHashing(t *testing.T) {
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if hash == testPassword {
		t.Error("password was not hashed")
	}
	testHash = hash

}

func TestComparing(t *testing.T) {
	if err := CheckPasswordHash(testPassword, testHash); err != nil {
		t.Error("hashes do not match")
	}
}

func TestCreatingAJWT(t *testing.T) {
	testUUID = uuid.New()
	token, err := MakeJWT(testUUID, tokenSecret, time.Second*5)
	if err != nil {
		t.Errorf("error creating token %v", err)
	}
	testToken = token
}

func TestParsingJWT(t *testing.T) {
	userID, err := ValidateJWT(testToken, tokenSecret)
	if err != nil {
		t.Errorf("error validating token %v", err)
	}
	if userID != testUUID {
		t.Errorf("generated user id %v doesn't equal the value in the jwt: %v", testUUID, userID)
	}
}

func TestRejectedOnExpiry(t *testing.T) {
	token, _ := MakeJWT(testUUID, tokenSecret, time.Second*1)
	time.Sleep(time.Second * 1)
	userID, err := ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Errorf("token validated after expiry, userID: %v", userID)
	}
}

func TestRejectedOnWrongSecret(t *testing.T) {
	otherSecret := "otherSecret"
	token, _ := MakeJWT(testUUID, otherSecret, time.Second*5)
	userID, err := ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Errorf("token validated with another secret, userID: %v", userID)
	}
}

func TestGetBearerToken(t *testing.T) {
	type testCase struct {
		Name        string
		Headers     http.Header
		Token       string
		ExpectError bool
	}
	testCases := []testCase{
		{
			Name: "Header with token",
			Headers: http.Header{
				http.CanonicalHeaderKey("Authorization"): []string{"Bearer theGoatOfKeys"},
			},
			Token:       "theGoatOfKeys",
			ExpectError: false,
		},
		{
			Name:        "Missing auth header",
			Headers:     http.Header{},
			Token:       "",
			ExpectError: true,
		},
		{
			Name: "Header missing token",
			Headers: http.Header{
				http.CanonicalHeaderKey("Authorization"): []string{"Bearer"},
			},
			Token:       "",
			ExpectError: true,
		},
	}
	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			token, err := GetBearerToken(c.Headers)
			if (err != nil) != c.ExpectError {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, c.ExpectError)
			}
			if token != c.Token {
				t.Errorf("Token %s does not equal original token %s", token, c.Token)
			}
		})
	}

}
