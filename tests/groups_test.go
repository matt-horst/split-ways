package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/mux"
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
	const username = "user"
	const password = "password"

	body, err := json.Marshal(handlers.CreateUserData{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	require.Equal(t, http.StatusCreated, rr.Code)

	user := handlers.ExportUser{}

	err = json.NewDecoder(rr.Result().Body).Decode(&user)
	require.NoError(t, err)

	// Login user
	body, err = json.Marshal(handlers.LoginUserData{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	r = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerLogin(rr, r)

	require.Equal(t, http.StatusNoContent, rr.Code)
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)

	// Successfully create new group
	const groupName = "New Group"

	body, err = json.Marshal(handlers.CreateGroupData{Name: groupName})
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

func TestUpdateGroup(t *testing.T) {
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
	const username = "user"
	const password = "password"

	body, err := json.Marshal(handlers.CreateUserData{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	cfg.HandlerCreateUser(rr, r)

	require.Equal(t, http.StatusCreated, rr.Code)

	user := handlers.ExportUser{}

	err = json.NewDecoder(rr.Result().Body).Decode(&user)
	require.NoError(t, err)

	// Login user
	body, err = json.Marshal(handlers.LoginUserData{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	r = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()

	cfg.HandlerLogin(rr, r)

	require.Equal(t, http.StatusNoContent, rr.Code)
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)
	jwt := cookies[0]

	// Create new group
	const groupName = "New Group"

	body, err = json.Marshal(handlers.CreateGroupData{Name: groupName})
	require.NoError(t, err)

	r = httptest.NewRequest("POST", "/api/groups", bytes.NewBuffer(body))
	r.AddCookie(jwt)
	rr = httptest.NewRecorder()

	cfg.AuthenticatedUserMiddleware(
		http.HandlerFunc(cfg.HandlerCreateGroup),
	).ServeHTTP(rr, r)

	require.Equal(t, http.StatusCreated, rr.Code)

	group := database.Group{}

	err = json.NewDecoder(rr.Result().Body).Decode(&group)
	require.NoError(t, err)

	// Test successful update of group
	const newGroupName = "Updated Group"
	body, err = json.Marshal(handlers.UpdateGroupData{
		Name: newGroupName,
	})

	r = httptest.NewRequest("PUT", "/api/groups/"+group.ID.String(), bytes.NewBuffer(body))
	r.AddCookie(jwt)
	r = mux.SetURLVars(r, map[string]string{"group_id": group.ID.String()})
	rr = httptest.NewRecorder()

	cfg.AuthenticatedUserMiddleware(
		http.HandlerFunc(cfg.HandlerUpdateGroup),
	).ServeHTTP(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	newGroup := database.Group{}
	err = json.NewDecoder(rr.Result().Body).Decode(&newGroup)
	assert.NoError(t, err)

	assert.Equal(t, newGroup.Name, newGroupName)
	assert.Equal(t, group.ID, newGroup.ID)
	assert.Equal(t, group.CreatedAt, newGroup.CreatedAt)
	assert.Equal(t, user.ID, newGroup.Owner)
}
