package repositories

import (
	"database/sql"
	"strconv"

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

func (r *EventRepository) InsertEvent(db *sql.DB, event models.Event) error {
	query := `
        INSERT INTO Events (
            id, title, room, tags, startTime, endTime, type, location, date, githubRepo, 
            slidesURL, imageSrc, virtualURL, extraImageSrcs, description, about, 
            attendeeIds, organizerIds, usersAttendedIds, createdAt, updatedAt, createdBy
        ) VALUES (
            $1, $2, ROW($3, $4, $5, $6), $7, $8, $9, $10, $11, $12, $13, 
            $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25
        )
    `

	_, err := db.Exec(query,
		event.ID, event.Title,
		// Room fields expanded
		event.Room.Building, event.Room.Room, event.Room.Type, event.Room.Capacity,
		pq.Array(event.Tags), event.StartTime, event.EndTime,
		event.Type, event.Location, event.Date, event.GithubRepo, event.SlidesURL, event.ImageSrc,
		event.VirtualURL, pq.Array(event.ExtraImageSrcs), event.Description, event.About,
		pq.Array(event.AttendeeIds), pq.Array(event.OrganizerIds), pq.Array(event.UsersAttendedIds),
		event.CreatedAt, event.UpdatedAt, event.CreatedBy,
	)

	return err
}

func (r *EventRepository) GetByID(id uuid.UUID) (*models.Event, error) {
	event := &models.Event{}
	query := `
		SELECT id, 
			title, 
			(room).building, 
			(room).room, 
			(room).type, 
			(room).capacity, 
			tags, 
			startTime, 
			endTime, 
			type, 
			location, 
			date, 
			githubRepo, 
			slidesURL, 
			imageSrc, 
			virtualURL, 
			extraImageSrcs, 
			description, 
			about, 
			attendeeIds, 
			organizerIds, 
			usersAttendedIds, 
			createdAt, 
			updatedAt, 
			createdBy
		FROM events
		WHERE id = $1
	`
	var building sql.NullString
	var capacity, room, roomType sql.NullInt32

	err := r.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.Title,
		&building,
		&room,
		&roomType,
		&capacity,
		pq.Array(&event.Tags),
		&event.StartTime,
		&event.EndTime,
		&event.Type,
		&event.Location,
		&event.Date,
		&event.GithubRepo,
		&event.SlidesURL,
		&event.ImageSrc,
		&event.VirtualURL,
		pq.Array(&event.ExtraImageSrcs),
		&event.Description,
		&event.About,
		pq.Array(&event.AttendeeIds),
		pq.Array(&event.OrganizerIds),
		pq.Array(&event.UsersAttendedIds),
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	event.Room = &models.CSUSMRoom{
		Building: building.String,
		Room:     int(room.Int32),
		Type:     models.RoomType(int(roomType.Int32)),
		Capacity: int(capacity.Int32),
	}

	return event, nil
}

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
			(room).building, 
			(room).room, 
			(room).type, 
			(room).capacity,
			tags, 
			startTime, 
			endTime, 
			type, 
			location, 
			date, 
			githubRepo, 
			slidesURL, 
			imageSrc, 
			virtualURL, 
			extraImageSrcs, 
			description, 
			about, 
			attendeeIds, 
			organizerIds, 
			usersAttendedIds, 
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

	var building sql.NullString
	var capacity, room, roomType sql.NullInt32

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&building,
			&room,
			&roomType,
			&capacity,
			pq.Array(&event.Tags),
			&event.StartTime,
			&event.EndTime,
			&event.Type,
			&event.Location,
			&event.Date,
			&event.GithubRepo,
			&event.SlidesURL,
			&event.ImageSrc,
			&event.VirtualURL,
			pq.Array(&event.ExtraImageSrcs),
			&event.Description,
			&event.About,
			pq.Array(&event.AttendeeIds),
			pq.Array(&event.OrganizerIds),
			pq.Array(&event.UsersAttendedIds),
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		event.Room = &models.CSUSMRoom{
			Building: building.String,
			Room:     int(room.Int32),
			Type:     models.RoomType(int(roomType.Int32)),
			Capacity: int(capacity.Int32),
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

func (r *EventRepository) GetTotalCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
