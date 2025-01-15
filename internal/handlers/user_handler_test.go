package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/csusmGDSC/csusmgdsc-api/internal/handlers"
	"github.com/csusmGDSC/csusmgdsc-api/internal/testutils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUser(t *testing.T) {
	// Setup test helper
	th := testutils.SetupIntegrationTest(t)
	defer th.Cleanup()

	// Create Echo instance
	e := echo.New()
	h := handlers.NewHandler(th.RealDB)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid Registration",
			requestBody: `{
				"email": "john.doe@csusm.edu",
				"first_name": "John",
				"last_name": "Doe",
				"password": "SecurePass123!",
				"role": "USER",
				"position": 1,
				"branch": 1,
				"graduation_date": "2024-05-15T00:00:00Z"
			}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid Email Format",
			requestBody: `{
				"email": "invalid-email",
				"first_name": "John",
				"last_name": "Doe",
				"password": "SecurePass123!",
				"role": "USER",
				"position": 1,
				"branch": 1,
				"graduation_date": "2024-05-15T00:00:00Z"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email",
		},
		{
			name: "Missing Required Fields",
			requestBody: `{
				"email": "john.doe@csusm.edu",
				"first_name": "John",
				"last_name": "Doe",
				"password": "SecurePass123!"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Role required",
		},
		{
			name: "Empty Password",
			requestBody: `{
                "email": "john.doe@csusm.edu",
				"first_name": "John",
				"last_name": "Doe",
				"password": "",
				"role": "USER",
				"position": 1,
				"branch": 1,
				"graduation_date": "2024-05-15T00:00:00Z"
            }`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup request
			req := httptest.NewRequest(http.MethodPost, "/auth/register",
				strings.NewReader(tc.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			// Setup recorder
			rec := httptest.NewRecorder()

			// Setup context
			c := e.NewContext(req, rec)

			// Perform request, evaluate response based on recorder rather than error
			_ = h.RegisterUser(c)

			// Test Cases that we expect to fail
			if tc.expectedStatus >= 400 {
				assert.Equal(t, tc.expectedStatus, rec.Code)
				if tc.expectedError != "" {
					assert.Contains(t, rec.Body.String(), tc.expectedError)
				}
				// Test Cases that we expect to pass
			} else if tc.expectedStatus == http.StatusCreated {
				assert.Equal(t, tc.expectedStatus, rec.Code)
			}
			th.CleanupDatabase()
		})
	}

	// Test duplicate email registration
	t.Run("Duplicate Email Registration", func(t *testing.T) {
		requestBody := `{
            "email": "john.doe@csusm.edu",
			"first_name": "John",
			"last_name": "Doe",
			"password": "SecurePass123!",
			"role": "USER",
			"position": 1,
			"branch": 1,
			"graduation_date": "2024-05-15T00:00:00Z"
        }`

		// First request to create the user
		req := httptest.NewRequest(http.MethodPost, "/auth/register",
			strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = h.RegisterUser(c)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Second request with same email
		req = httptest.NewRequest(http.MethodPost, "/auth/register",
			strings.NewReader(requestBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		_ = h.RegisterUser(c)

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "Email already registered")
		th.CleanupDatabase()
	})
}
