package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) getComments(query string, args ...interface{}) ([]*models.Comment, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying comments: %v", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		var commentIds, likes, dislikes pq.StringArray
		err := rows.Scan(&comment.ID, &comment.UserId, &comment.EventId, &comment.Content, &comment.PinnedBy, &commentIds, &comment.CreatedAt, &comment.UpdatedAt, &likes, &dislikes)
		if err != nil {
			return nil, fmt.Errorf("error scanning comment: %v", err)
		}

		comment.CommentIds = make([]uuid.UUID, len(commentIds))
		for i, id := range commentIds {
			comment.CommentIds[i], err = uuid.Parse(id)
			if err != nil {
				return nil, fmt.Errorf("error parsing commentId: %v", err)
			}
		}

		comment.Likes = make([]uuid.UUID, len(likes))
		for i, id := range likes {
			comment.Likes[i], err = uuid.Parse(id)
			if err != nil {
				return nil, fmt.Errorf("error parsing like: %v", err)
			}
		}

		comment.Dislikes = make([]uuid.UUID, len(dislikes))
		for i, id := range dislikes {
			comment.Dislikes[i], err = uuid.Parse(id)
			if err != nil {
				return nil, fmt.Errorf("error parsing dislike: %v", err)
			}
		}

		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *CommentRepository) CreateComment(db *sql.DB, comment models.Comment) error {
	query := `
		INSERT INTO Comments (
			id, userId, eventId, content, pinnedBy, commentIds, createdAt, updatedAt, likes, dislikes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		);
    `

	_, err := r.db.Exec(query, comment.ID, comment.UserId, comment.EventId, comment.Content, comment.PinnedBy,
		pq.Array(comment.CommentIds), comment.CreatedAt, comment.UpdatedAt, pq.Array(comment.Likes), pq.Array(comment.Dislikes))

	if err != nil {
		return fmt.Errorf("error inserting comment: %v", err)
	}

	return nil
}

func (r *CommentRepository) DeleteCommentById(id uuid.UUID) error {
	query := `DELETE FROM Comments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *CommentRepository) GetCommentsByUserId(userId uuid.UUID) ([]*models.Comment, error) {
	return r.getComments(`SELECT * FROM Comments WHERE userId = $1`, userId)
}

func (r *CommentRepository) GetCommentsByEventId(eventId uuid.UUID) ([]*models.Comment, error) {
	return r.getComments(`SELECT * FROM Comments WHERE eventId = $1`, eventId)
}

func (r *CommentRepository) GetCommentByCommentId(id uuid.UUID) (*models.Comment, error) {
	comments, _ := r.getComments(`SELECT * FROM Comments WHERE id = $1`, id)
	return comments[0], nil // only one comment should be returned
}

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
		updates = append(updates, fmt.Sprintf("pinnedBy = $%d", valueIndex))
		values = append(values, *comment.PinnedBy) // Dereference the pointer
		valueIndex++
	}

	if comment.CommentIds != nil { // Check for pointer or zero value if appropriate
		updates = append(updates, fmt.Sprintf("commentIds = $%d", valueIndex))
		values = append(values, pq.Array(comment.CommentIds)) // Dereference and use pq.Array
		valueIndex++
	}

	if comment.Likes != nil { // Check for pointer or zero value if appropriate
		updates = append(updates, fmt.Sprintf("likes = $%d", valueIndex))
		values = append(values, pq.Array(comment.Likes)) // Dereference and use pq.Array
		valueIndex++
	}

	if comment.Dislikes != nil { // Check for pointer or zero value if appropriate
		updates = append(updates, fmt.Sprintf("dislikes = $%d", valueIndex))
		values = append(values, pq.Array(comment.Dislikes)) // Dereference and use pq.Array
		valueIndex++
	}

	updates = append(updates, fmt.Sprintf("updatedAt = $%d", valueIndex)) // Always update updatedAt
	values = append(values, time.Now())
	valueIndex++

	if len(updates) == 0 {
		return nil // Or an error, if you require at least one field to be updated.
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
