package stormpath

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIKey(t *testing.T) {
	t.Parallel()

	application := createTestApplication(t)
	defer application.Purge()

	account := createTestAccount(application, t)
	key := APIKey{
		Name: "test",
	}
	apiKey, _ := account.CreateAPIKey(&key)

	k, err := GetAPIKey(apiKey.Href, MakeAPIKeyCriteria())

	assert.NoError(t, err)
	assert.Equal(t, apiKey, k)
}

func TestDeleteAPIKey(t *testing.T) {
	t.Parallel()

	application := createTestApplication(t)
	defer application.Purge()

	account := createTestAccount(application, t)
	key := APIKey{}
	apiKey, _ := account.CreateAPIKey(&key)

	err := apiKey.Delete()

	assert.NoError(t, err)

	k, err := GetAPIKey(apiKey.Href, MakeAPIKeyCriteria())

	assert.Error(t, err)
	assert.Nil(t, k)
	assert.Equal(t, http.StatusNotFound, err.(Error).Status)
}

func TestUpdateAPIKey(t *testing.T) {
	t.Parallel()

	application := createTestApplication(t)
	defer application.Purge()

	account := createTestAccount(application, t)
	key := APIKey{}
	apiKey, _ := account.CreateAPIKey(&key)

	apiKey.Status = Disabled
	apiKey.Name = "Test"
	err := apiKey.Update()

	assert.NoError(t, err)

	updatedAPIKey, _ := GetAPIKey(apiKey.Href, MakeAPIKeyCriteria())
	assert.Equal(t, Disabled, updatedAPIKey.Status)
	assert.Equal(t, "Test", updatedAPIKey.Name)
}

func TestGetAPIKeys(t *testing.T) {
	t.Parallel()

	application := createTestApplication(t)
	defer application.Purge()

	account := createTestAccount(application, t)
	ak1 := APIKey{}
	ak2 := APIKey{}
	apiKey1, _ := account.CreateAPIKey(&ak1)
	apiKey2, _ := account.CreateAPIKey(&ak2)

	keys, err := GetAPIKeys(account.APIKeys.Href, MakeAPIKeyCriteria())

	assert.NoError(t, err)
	assert.Equal(t, apiKey1, &keys.Items[0])
	assert.Equal(t, apiKey2, &keys.Items[1])
}