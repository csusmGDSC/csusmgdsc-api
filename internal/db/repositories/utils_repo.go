package repositories

import (
	"database/sql"

	"github.com/google/uuid"
)

type UtilsRepository struct {
	db *sql.DB
}

func NewUtilsRepository(db *sql.DB) *UtilsRepository {
	return &UtilsRepository{db: db}
}

// CheckIfUUIDExists checks if a UUID exists in a specified table and column within the database.
//
// Parameters:
//   - table: the name of the table to search in.
//   - column: the name of the column to search within.
//   - id: the UUID to check for existence.
//
// Returns:
//   - bool: true if the UUID exists in the specified table and column, otherwise false.
//   - error: an error if there is an issue querying the database.
func (r *UtilsRepository) CheckIfUUIDExists(table string, column string, id uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM " + table + " WHERE " + column + " = $1)"
	var exists bool
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
