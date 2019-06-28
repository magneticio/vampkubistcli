package client_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryTokenStore(t *testing.T) {
	var tokenStore client.TokenStore
	tokenStore = &client.InMemoryTokenStore{}
	token := "test-token"
	timeout := int64(5)
	storeError := tokenStore.Store(token, timeout)
	assert.Equal(t, nil, storeError)
	val, ok := tokenStore.Get(token)
	assert.Equal(t, true, ok)
	assert.Equal(t, timeout, val)
	expetedTokenMap := map[string]int64{token: timeout}
	tokenMap := tokenStore.Tokens()
	assert.Equal(t, expetedTokenMap, tokenMap)
	removeError := tokenStore.RemoveExpired()
	assert.Equal(t, nil, removeError)
}

func TestFileBackedTokenStore(t *testing.T) {
	var tokenStore client.TokenStore
	tmpfile, err := ioutil.TempFile("", "tokenstore")
	assert.Equal(t, nil, err)
	defer os.Remove(tmpfile.Name()) // clean up

	tokenStore = &client.FileBackedTokenStore{
		Path: tmpfile.Name(),
	}
	token := "test-token"
	timeout := int64(5)
	storeError := tokenStore.Store(token, timeout)
	assert.Equal(t, nil, storeError)
	val, ok := tokenStore.Get(token)
	assert.Equal(t, true, ok)
	assert.Equal(t, timeout, val)
	expetedTokenMap := map[string]int64{token: timeout}
	tokenMap := tokenStore.Tokens()
	assert.Equal(t, expetedTokenMap, tokenMap)
	removeError := tokenStore.RemoveExpired()
	assert.Equal(t, nil, removeError)
}
