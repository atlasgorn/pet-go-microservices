package aaa_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/api/adapters/aaa"
)

func TestNew_MissingEnvVars(t *testing.T) {
	auth, err := aaa.New(time.Hour, nil)

	assert.Error(t, err)
	assert.Equal(t, aaa.AAA{}, auth)
}

func TestNew_Success(t *testing.T) {
	t.Setenv("ADMIN_USER", "testadmin")
	t.Setenv("ADMIN_PASSWORD", "testpass")

	auth, err := aaa.New(time.Hour, nil)

	require.NoError(t, err)
	assert.NotNil(t, auth)
}

func TestAAA_Login_Success(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "secret")

	auth, _ := aaa.New(time.Hour, nil)

	token, err := auth.Login("admin", "secret")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAAA_Login_WrongUser(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "secret")

	auth, _ := aaa.New(time.Hour, nil)

	token, err := auth.Login("wronguser", "secret")

	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAAA_Login_WrongPassword(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "secret")

	auth, _ := aaa.New(time.Hour, nil)

	token, err := auth.Login("admin", "wrongpass")

	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAAA_Verify_ValidToken(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "secret")

	auth, _ := aaa.New(time.Hour, nil)
	token, _ := auth.Login("admin", "secret")

	err := auth.Verify(token)

	assert.NoError(t, err)
}

func TestAAA_Verify_InvalidToken(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "secret")

	auth, _ := aaa.New(time.Hour, nil)

	err := auth.Verify("invalid.token.here")

	assert.Error(t, err)
}

func TestNew_MissingAdminPassword(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")

	auth, err := aaa.New(time.Hour, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin password")
	assert.Equal(t, aaa.AAA{}, auth)
}

// Test Verify with a valid token but subject != "superuser"
func TestAAA_Verify_WrongSubject(t *testing.T) {
	const secretKey = "something secret here"

	claims := jwt.RegisteredClaims{
		Subject:   "regular_user",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// Create AAA instance (needed for Verify, though only the secret matters)
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "pass")
	auth, _ := aaa.New(time.Hour, nil)

	err = auth.Verify(signed)
	assert.Error(t, err)
	assert.EqualError(t, err, "token subject is not superuser")
}

// Test Verify with a token missing the "sub" claim
func TestAAA_Verify_MissingSubject(t *testing.T) {
	const secretKey = "something secret here"

	claims := jwt.RegisteredClaims{
		// No Subject field
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "pass")
	auth, _ := aaa.New(time.Hour, nil)

	err = auth.Verify(signed)
	assert.Error(t, err)
	assert.EqualError(t, err, "token subject is not superuser")
}
