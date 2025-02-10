package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// getComments queries the database for comments using the given query and arguments.
// The function returns a slice of pointers to Comment objects and an error.
// The error is non-nil if there was an error querying the database or scanning the results.
func (r *CommentRepository) getComments(query string, args ...interface{}) ([]*models.Comment, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying comments: %v", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.UserId, &comment.EventId, &comment.Content, &comment.PinnedBy, &comment.CreatedAt, &comment.UpdatedAt, &comment.ParentId)
		if err != nil {
			return nil, fmt.Errorf("error scanning comment: %v", err)
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

// CreateComment inserts a new comment into the database.
//
// The function takes a pointer to a database connection and a pointer to a Comment
// object as arguments. The function returns an error if there is an error inserting
// the comment into the database.
func (r *CommentRepository) CreateComment(db *sql.DB, comment models.Comment) error {
	query := `
		INSERT INTO Comments (
			id, user_id, event_id, content, pinned_by, created_at, updated_at, parent_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		);
    `

	_, err := r.db.Exec(query, comment.ID, comment.UserId, comment.EventId, comment.Content, comment.PinnedBy,
		comment.CreatedAt, comment.UpdatedAt, comment.ParentId)

	if err != nil {
		return fmt.Errorf("error inserting comment: %v", err)
	}

	return nil
}

// DeleteCommentById deletes a comment from the database by its ID.
//
// The function takes a UUID and executes a DELETE query on the Comments table.
// If there is an error executing the query, the error is returned.
func (r *CommentRepository) DeleteCommentById(id uuid.UUID) error {
	query := `DELETE FROM Comments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// GetAllComments returns a slice of pointers to Comment objects that represent all comments in the database.
// The function returns an error if there is an error querying the database or scanning the results.
func (r *CommentRepository) GetAllComments() ([]*models.Comment, error) {
	return r.getComments(`SELECT * FROM Comments`)
}

// GetCommentsByUserId returns a slice of pointers to Comment objects that represent comments posted by the user with the given ID.
// The function returns an error if there is an error querying the database or scanning the results.
func (r *CommentRepository) GetCommentsByUserId(userId uuid.UUID) ([]*models.Comment, error) {
	return r.getComments(`SELECT * FROM Comments WHERE user_id = $1`, userId)
}

// GetCommentsWithRepliesByEventId returns a slice of pointers to Comment objects that represent
// all comments and their replies for the given event ID.
//
// The function uses a recursive common table expression (CTE) to fetch all comments and their replies.
// Comments are stored in a map by ID, and the reply tree is reconstructed from the map.
// The function returns an error if there is an error querying the database or scanning the results.
func (r *CommentRepository) GetCommentsWithRepliesByEventId(eventId uuid.UUID) ([]*models.Comment, error) {
	query := `
        WITH RECURSIVE CommentTree AS (
            SELECT id, user_id, event_id, content, pinned_by, created_at, updated_at, parent_id
            FROM Comments
            WHERE event_id = $1 AND parent_id IS NULL -- Start with top-level comments

            UNION ALL

            SELECT c.id, c.user_id, c.event_id, c.content, c.pinned_by, c.created_at, c.updated_at, c.parent_id
            FROM Comments c
            INNER JOIN CommentTree ct ON c.parent_id = ct.id
        )
        SELECT * FROM CommentTree ORDER BY created_at;
    `

	rows, err := r.db.Query(query, eventId)
	if err != nil {
		return nil, fmt.Errorf("error querying comments with replies: %v", err)
	}
	defer rows.Close()

	commentMap := make(map[uuid.UUID]*models.Comment) // Store comments by ID
	var comments []*models.Comment

	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.UserId, &comment.EventId, &comment.Content, &comment.PinnedBy, &comment.CreatedAt, &comment.UpdatedAt, &comment.ParentId)
		if err != nil {
			return nil, fmt.Errorf("error scanning comment: %v", err)
		}

		commentMap[comment.ID] = &comment // Add to map

		if comment.ParentId == nil { // Top-level comment
			comments = append(comments, &comment)
		} else { // Reply
			if parent, ok := commentMap[*comment.ParentId]; ok {
				if parent.Replies == nil {
					parent.Replies = make([]*models.Comment, 0) // Initialize if not present
				}
				parent.Replies = append(parent.Replies, &comment)
			}
		}
	}

	return comments, nil
}

// GetCommentsByUserIdAndEventId returns a slice of pointers to Comment objects that represent comments posted by the user with the given ID
// and referring to the event with the given ID.
//
// The function returns an error if there is an error querying the database or scanning the results.
func (r *CommentRepository) GetCommentsByUserIdAndEventId(userId uuid.UUID, eventId uuid.UUID) ([]*models.Comment, error) {
	return r.getComments(`SELECT * FROM Comments WHERE user_id = $1 AND event_id = $2`, userId, eventId)
}

// GetCommentByCommentId retrieves a comment by its unique identifier.
//
// The function takes a UUID representing the comment ID as an argument
// and returns a pointer to a Comment object and an error. The error is
// non-nil if there was an issue querying the database or if the comment
// could not be found. Only one comment should be returned for the given
// ID.
func (r *CommentRepository) GetCommentByCommentId(id uuid.UUID) (*models.Comment, error) {
	comments, _ := r.getComments(`SELECT * FROM Comments WHERE id = $1`, id)
	return comments[0], nil // only one comment should be returned
}

// UpdateCommentByCommentId updates the fields of a comment in the database by its ID.
//
// The function takes a UUID representing the comment ID and an UpdateCommentRequest object
// containing the fields to be updated. The fields that can be updated include:
// - content: the updated text of the comment
// - pinned_by: the ID of the user who pinned the comment
// - parent_id: the ID of the comment's parent
//
// The function constructs an SQL UPDATE query dynamically based on the fields provided in
// the UpdateCommentRequest object. It also updates the "updated_at" timestamp to the current time.
//
// If no fields are provided for update, the function returns early with no error.
// The function returns an error if there is an issue executing the update query.
func (r *CommentRepository) UpdateCommentByCommentId(id uuid.UUID, comment models.UpdateCommentRquest) error {
	updates := make([]string, 0)
	values := make([]interface{}, 0)
	valueIndex := 1 // Start at 1 because $1 is commentID

	if comment.Content != nil { // Check for pointer or zero value if appropriate
		updates = append(updates, fmt.Sprintf("content = $%d", valueIndex))
		values = append(values, *comment.Content) // Dereference the pointer
		valueIndex++
	}

	if comment.PinnedBy != nil { // Check for pointer or zero value if appropriate
		updates = append(updates, fmt.Sprintf("pinned_by = $%d", valueIndex))
		values = append(values, *comment.PinnedBy) // Dereference the pointer
		valueIndex++
	}

	if comment.ParentId != nil { // Check for pointer or zero value if appropriate
		updates = append(updates, fmt.Sprintf("parent_id = $%d", valueIndex))
		values = append(values, *comment.ParentId) // Dereference the pointer
		valueIndex++
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", valueIndex)) // Always update updatedAt
	values = append(values, time.Now())
	valueIndex++

	if len(updates) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
        UPDATE Comments
        SET %s
        WHERE id = $%d`,
		strings.Join(updates, ", "), valueIndex)

	values = append(values, id)

	_, err := r.db.Exec(query, values...)
	return err
}

// GetRepliesByCommentId retrieves all comments that are replies to the comment with the given ID.
//
// The function takes a UUID representing the comment ID as an argument and returns a slice of pointers
// to Comment objects and an error. The error is non-nil if there is an issue querying the database or
// scanning the results.
func (r *CommentRepository) GetRepliesByCommentId(commentId uuid.UUID) ([]*models.Comment, error) {
	return r.getComments(`SELECT * FROM Comments WHERE parent_id = $1`, commentId)
}
