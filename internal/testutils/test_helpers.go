package testutils

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/csusmGDSC/csusmgdsc-api/internal/handlers"
	"github.com/csusmGDSC/csusmgdsc-api/internal/mocks"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type TestHelper struct {
	T       *testing.T
	Ctrl    *gomock.Controller
	MockDB  *mocks.MockDatabaseConnection
	RealDB  db.DatabaseConnection
	Handler *handlers.Handler
	Echo    *echo.Echo
	Cleanup func()
}

// SetupTest creates a new test helper with mocked dependencies
func SetupTest(t *testing.T) *TestHelper {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDatabaseConnection(ctrl)
	e := echo.New()
	h := handlers.NewHandler(mockDB)

	return &TestHelper{
		T:       t,
		Ctrl:    ctrl,
		MockDB:  mockDB,
		Echo:    e,
		Handler: h,
		Cleanup: func() {
			ctrl.Finish()
		},
	}
}

// SetupIntegrationTest creates a new test helper with real dependencies
func SetupIntegrationTest(t *testing.T) *TestHelper {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Set test environment
	t.Setenv("GO_ENV", "test")

	// Print current working directory for debugging
	dir, _ := os.Getwd()
	t.Logf("Current working directory: %s", dir)

	// Setup real database connection
	conn := db.NewPostgresConnection()
	err := conn.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	e := echo.New()
	h := handlers.NewHandler(conn)

	return &TestHelper{
		T:       t,
		RealDB:  conn,
		Echo:    e,
		Handler: h,
		Cleanup: func() {
			if err := conn.Close(); err != nil {
				t.Errorf("Failed to close test database connection: %v", err)
			}
		},
	}
}

// CreateTestContext creates a new echo.Context for testing
func (th *TestHelper) CreateTestContext(method, path string, body interface{}) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			th.T.Fatalf("Failed to marshal request body: %v", err)
		}
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	c := th.Echo.NewContext(req, rec)
	return c, rec
}

// SetAuthenticatedUser sets up the context for an authenticated user
func (th *TestHelper) SetAuthenticatedUser(c echo.Context, userID string, role models.Role) {
	c.Set("user_id", userID)
	c.Set("user_role", role)
}

// AssertJSONResponse helps verify JSON responses
func (th *TestHelper) AssertJSONResponse(rec *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	assert.Equal(th.T, expectedStatus, rec.Code)

	var actualBody interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &actualBody)
	assert.NoError(th.T, err)
	assert.Equal(th.T, expectedBody, actualBody)
}

// AssertErrorResponse helps verify error responses
func (th *TestHelper) AssertErrorResponse(rec *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	assert.Equal(th.T, expectedStatus, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(th.T, err)
	assert.Equal(th.T, expectedError, response["error"])
}

// CleanupDatabase cleans up the database for integration tests
func (th *TestHelper) CleanupDatabase() {
	if th.RealDB != nil {
		db := th.RealDB.GetDB()
		// Clean up tables in reverse order of dependencies
		db.Exec("DELETE FROM events")
		db.Exec("DELETE FROM refresh_tokens")
		db.Exec("DELETE FROM users")
	}
}
