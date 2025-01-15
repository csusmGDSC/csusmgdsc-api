package db_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/csusmGDSC/csusmgdsc-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseConnection(t *testing.T) {
	th := testutils.SetupTest(t)
	defer th.Cleanup()

	tests := []struct {
		name    string
		setup   func()
		testFn  func() error
		wantErr bool
	}{
		{
			name: "successful connection",
			setup: func() {
				// Expect Connect to be called exactly once
				th.MockDB.EXPECT().
					Connect().
					Times(1).
					Return(nil)
			},
			testFn: func() error {
				return th.MockDB.Connect()
			},
			wantErr: false,
		},
		{
			name: "connection error",
			setup: func() {
				th.MockDB.EXPECT().
					Connect().
					Return(fmt.Errorf("connection failed"))
			},
			testFn: func() error {
				return th.MockDB.Connect()
			},
			wantErr: true,
		},
		{
			name: "successful close",
			setup: func() {
				th.MockDB.EXPECT().
					Close().
					Times(1).
					Return(nil)
			},
			testFn: func() error {
				return th.MockDB.Close()
			},
			wantErr: false,
		},
		{
			name: "close error",
			setup: func() {
				th.MockDB.EXPECT().
					Close().
					Return(fmt.Errorf("close failed"))
			},
			testFn: func() error {
				return th.MockDB.Close()
			},
			wantErr: true,
		},
		{
			name: "get database connection",
			setup: func() {
				mockDB := &sql.DB{} // Create a mock sql.DB
				th.MockDB.EXPECT().
					GetDB().
					Times(1).
					Return(mockDB)
			},
			testFn: func() error {
				db := th.MockDB.GetDB()
				if db == nil {
					return fmt.Errorf("expected non-nil database connection")
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setup()

			// Execute test
			err := tt.testFn()

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPostgresConnection(t *testing.T) {
	// Test object creation
	conn := db.NewPostgresConnection()
	assert.NotNil(t, conn, "Expected non-nil connection object")
	assert.Implements(t, (*db.DatabaseConnection)(nil), conn, "Expected connection to implement DatabaseConnection interface")

	// Test type assertion
	_, ok := conn.(*db.PostgresConnection)
	assert.True(t, ok, "Expected connection to be of type *PostgresConnection")

	// Test initial state
	pgConn, _ := conn.(*db.PostgresConnection)
	assert.Nil(t, pgConn.GetDB(), "Expected initial database connection to be nil")
}

func TestIntegrationConnection(t *testing.T) {
	start := time.Now()
	defer func() {
		t.Logf("Integration test took: %v", time.Since(start))
	}()

	th := testutils.SetupIntegrationTest(t)
	defer th.Cleanup()

	// Test database connection
	assert.NotNil(t, th.RealDB, "Expected non-nil RealDB")

	// Test getting database instance
	dbConn := th.RealDB.GetDB()
	assert.NotNil(t, dbConn, "Expected non-nil database connection")

	// Test actual database connection
	err := dbConn.Ping()
	assert.NoError(t, err, "Expected successful database ping")

	// Test queries
	_, err = dbConn.Exec("SELECT 1")
	assert.NoError(t, err, "Expected successful query execution")

	// Optional: Clean up any test data
	th.CleanupDatabase()
}
