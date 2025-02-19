package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// InsertEvent inserts a new event record into the events table in the database.
//
// It takes a database connection and an Event object as parameters. The room field
// of the Event object is converted to JSONB format before insertion.
//
// It returns an error if the room conversion to JSONB fails or if the database
// insertion fails.
func (r *EventRepository) InsertEvent(db *sql.DB, event models.Event) error {
	// Convert the room struct to JSONB format
	roomJSON, err := json.Marshal(event.Room)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO events (
            id, title, room, tags, start_time, end_time, type, location, date, repository_url, 
            slides_url, image_src, virtual_url, description, about, created_at, updated_at, created_by
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
            $11, $12, $13, $14, $15, $16, $17, $18
        );
    `

	_, err = db.Exec(query,
		uuid.New(),
		event.Title,
		roomJSON,
		pq.Array(event.Tags),
		event.StartTime,
		event.EndTime,
		event.Type,
		event.Location,
		event.Date,
		event.RepositoryURL,
		event.SlidesURL,
		event.ImageSrc,
		event.VirtualURL,
		event.Description,
		event.About,
		time.Now(),
		time.Now(),
		event.CreatedBy,
	)

	return err
}

// GetByID retrieves an event given its ID.
//
// It queries the events table and returns an Event object that corresponds to the given ID.
//
// The function returns an error if the query fails.
func (r *EventRepository) GetByID(id uuid.UUID) (*models.Event, error) {
	event := &models.Event{}
	query := `
		SELECT id, title, room, tags, start_time, end_time, type, location, date, repository_url, 
		    slides_url, image_src, virtual_url, description, about, created_at, updated_at, created_by
		FROM events
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.Title,
		&event.Room,
		pq.Array(&event.Tags),
		&event.StartTime,
		&event.EndTime,
		&event.Type,
		&event.Location,
		&event.Date,
		&event.RepositoryURL,
		&event.SlidesURL,
		&event.ImageSrc,
		&event.VirtualURL,
		&event.Description,
		&event.About,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetAll retrieves a paginated list of events from the database.
//
// It takes pagination parameters as string inputs for the page number and limit,
// converting them to integers with default values if necessary. It calculates the
// offset based on the page and limit, queries the events table, and returns an
// AllEventsResponse containing the list of Event objects, total count of events,
// current page, and limit.
//
// The function returns an error if the query or total count retrieval fails.
func (r *EventRepository) GetAll(pageStr string, limitStr string) (*models.AllEventsResponse, error) {
	// Convrt string parameters to integers with default values
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	//Calculate offset
	offset := (page - 1) * limit

	query := `
		SELECT 
			id, 
			title, 
			room,
			tags, 
			start_time, 
			end_time, 
			type, 
			location, 
			date, 
			repository_url, 
			slides_url, 
			imageSrc, 
			virtual_url, 
			description, 
			about, 
			createdAt, 
			updatedAt, 
			createdBy
		FROM events
		ORDER BY createdAt DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Room,
			pq.Array(&event.Tags),
			&event.StartTime,
			&event.EndTime,
			&event.Type,
			&event.Location,
			&event.Date,
			&event.RepositoryURL,
			&event.SlidesURL,
			&event.ImageSrc,
			&event.VirtualURL,
			&event.Description,
			&event.About,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, &event)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalCount, err := r.GetTotalCount()
	if err != nil {
		return nil, err
	}

	return &models.AllEventsResponse{
		Events:     events,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}, nil
}

// GetTotalCount returns the total count of events in the database.
//
// It queries the events table and returns the total count of events as an integer
// and an error if the query fails.
func (r *EventRepository) GetTotalCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteEventById deletes an event given its ID.
//
// It queries the events table and deletes the row with the given ID.
//
// The function returns an error if the query fails.
func (r *EventRepository) DeleteEventById(id uuid.UUID) error {
	query := "DELETE FROM events WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// UpdateEventById updates an event given its ID.
//
// It takes an EventID and an UpdateEventRequest object as parameters. The function
// updates the event record in the database with the fields that are not empty in the
// UpdateEventRequest object, and returns an error if the query fails.
//
// If no fields are changed, the function returns nil.
func (r *EventRepository) UpdateEventById(id uuid.UUID, event models.UpdateEventRequest) error {
	updates := make([]string, 0)
	values := make([]interface{}, 0)
	valueIndex := 1 // Start at 1 because $1 is EventID

	// Helper function to append fields to updates and values
	appendIfNotNil := func(fieldName string, fieldValue interface{}) {
		if fieldValue != nil {
			updates = append(updates, fmt.Sprintf("%s = $%d", fieldName, valueIndex))
			values = append(values, fieldValue)
			valueIndex++
		}
	}

	// Convert room struct into JSONB
	roomJSON, err := json.Marshal(event.Room)
	if err != nil {
		return err
	}

	appendIfNotNil("title", event.Title)
	appendIfNotNil("about", event.About)
	appendIfNotNil("date", event.Date)
	appendIfNotNil("description", event.Description)
	appendIfNotNil("start_time", event.StartTime)
	appendIfNotNil("end_time", event.EndTime)
	appendIfNotNil("image_src", event.ImageSrc)
	appendIfNotNil("location", event.Location)
	appendIfNotNil("repository_url", event.RepositoryURL)
	appendIfNotNil("virtual_url", event.VirtualURL)
	appendIfNotNil("slides_url", event.SlidesURL)
	appendIfNotNil("tags", pq.Array(event.Tags))
	appendIfNotNil("type", event.Type)
	appendIfNotNil("room", roomJSON)

	if len(updates) == 0 {
		return nil
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", valueIndex))
	values = append(values, time.Now())
	valueIndex++

	query := fmt.Sprintf(`
		UPDATE events
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), valueIndex)

	values = append(values, id)

	_, err = r.db.Exec(query, values)
	return err
}
