package repositories

import (
	"database/sql"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
)

type EventOrganizerRepository struct {
	db *sql.DB
}

func NewEventOrganizerRepository(db *sql.DB) *EventOrganizerRepository {
	return &EventOrganizerRepository{db: db}
}

// InsertEventOrganizer inserts a new event organizer into the database.
//
// It takes an EventOrganizer object as a parameter and inserts its event ID, user ID,
// and the current timestamp as the creation date into the event_organizers table.
//
// Returns an error if the insertion fails.
func (r *EventOrganizerRepository) InsertEventOrganizer(eventOrganizer models.EventOrganizer) error {
	query := `
		INSERT INTO event_organizers (event_id, user_id, created_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, eventOrganizer.EventID, eventOrganizer.UserID, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// GetEventOrganizers retrieves a list of users who are organizers of a given event ID.
//
// It queries the event_organizers table, joins it with the users table, and returns a list
// of User objects that correspond to the organizers of the event.
//
// The function returns an error if the query fails.
func (r *EventOrganizerRepository) GetEventOrganizers(eventID uuid.UUID) ([]models.User, error) {
	query := `
        SELECT
			u.id, 
			u.full_name,
			u.email,
			u.image,
			u.role,
			u.position,
			u.branch
        FROM event_organizers eo
        JOIN users u ON eo.user_id = u.id
        WHERE eo.event_id = $1
    `

	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.FullName, &user.Email, &user.Image, &user.Role, &user.Position, &user.Branch)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetEventsByUserID retrieves a list of events that a given user is an organizer of.
//
// It queries the event_organizers table, joins it with the events table, and returns a list
// of Event objects that the user is organizing.
//
// The function returns an error if the query fails.
func (r *EventOrganizerRepository) GetEventsByUserID(userID uuid.UUID) ([]models.Event, error) {
	query := `
        SELECT e.id, e.title, e.room, e.tags, e.start_time, e.end_time, e.type, e.location, 
               e.date, e.repository_url, e.slides_url, e.image_src, e.virtual_url, e.description, 
               e.about, e.created_at, e.updated_at, e.created_by
        FROM event_organizers eo
        JOIN events e ON eo.event_id = e.id
        WHERE eo.user_id = $1
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.Title, &event.Room, &event.Tags, &event.StartTime, &event.EndTime,
			&event.Type, &event.Location, &event.Date, &event.RepositoryURL, &event.SlidesURL,
			&event.ImageSrc, &event.VirtualURL, &event.Description, &event.About,
			&event.CreatedAt, &event.UpdatedAt, &event.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// DeleteEventOrganizer deletes an event organizer from the database. It takes two parameters, an event ID and a user ID, and deletes the row from the event_organizers table that matches these IDs.
func (r *EventOrganizerRepository) DeleteEventOrganizer(eventID, userID uuid.UUID) error {
	query := `
		DELETE FROM event_organizers
		WHERE event_id = $1 AND user_id = $2
	`
	_, err := r.db.Exec(query, eventID, userID)
	if err != nil {
		return err
	}
	return nil
}
