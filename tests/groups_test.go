package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/sessions"

	"github.com/matt-horst/split-ways/handlers"
	"github.com/matt-horst/split-ways/internal/database"
)

func TestCreateGroup(t *testing.T) {
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	cfg := &handlers.Config{
		DB:      db,
		Queries: queries.WithTx(tx),
		Store:   sessions.NewCookieStore([]byte(sessionKey)),
		JwtKey:  jwtKey,
	}

	// Create user
	type createUserData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	const username = "user"
	const password = "password"

	body, err := json.Marshal(createUserData{Username: username, Password: password})
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	require.Equal(t, http.StatusCreated, rr.Code)

	user := handlers.ExportUser{}

	err = json.NewDecoder(rr.Result().Body).Decode(&user)
	require.NoError(t, err)

	// Login user
	type loginUserData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	body, err = json.Marshal(loginUserData{Username: username, Password: password})
	require.NoError(t, err)

	r = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerLogin(rr, r)

	require.Equal(t, http.StatusNoContent, rr.Code)
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)

	// Successfully create new group
	type createGroupData struct {
		Name string `json:"name"`
	}

	const groupName = "New Group"

	body, err = json.Marshal(createGroupData{Name: groupName})
	require.NoError(t, err)

	r = httptest.NewRequest("POST", "/api/groups", bytes.NewBuffer(body))
	r.AddCookie(cookies[0])
	rr = httptest.NewRecorder()

	cfg.AuthenticatedUserMiddleware(
		http.HandlerFunc(cfg.HandlerCreateGroup),
	).ServeHTTP(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	group := database.Group{}

	err = json.NewDecoder(rr.Result().Body).Decode(&group)
	assert.NoError(t, err)

	assert.Equal(t, groupName, group.Name)
	assert.Equal(t, user.ID, group.Owner)
}
