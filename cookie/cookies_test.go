package cookie

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAndDecodeAccessToken(t *testing.T) {
	const secret = "0123456789abcdefghijklmnopqrstuv"
	const token = "my access token"
	c, err := NewCipher([]byte(secret))
	assert.Nil(t, err)

	encoded, err := c.Encrypt(token)
	assert.Nil(t, err)

	decoded, err := c.Decrypt(encoded)
	assert.Nil(t, err)

	assert.NotEqual(t, token, encoded)
	assert.Equal(t, token, decoded)
}

func TestEncodeAndDecodeAccessTokenB64(t *testing.T) {
	const secret_b64 = "A3Xbr6fu6Al0HkgrP1ztjb-mYiwmxgNPP-XbNsz1WBk="
	const token = "my access token"

	secret, err := base64.URLEncoding.DecodeString(secret_b64)
	assert.Nil(t, err)
	c, err := NewCipher(secret)
	assert.Nil(t, err)

	encoded, err := c.Encrypt(token)
	assert.Nil(t, err)

	decoded, err := c.Decrypt(encoded)
	assert.Nil(t, err)

	assert.NotEqual(t, token, encoded)
	assert.Equal(t, token, decoded)
}
