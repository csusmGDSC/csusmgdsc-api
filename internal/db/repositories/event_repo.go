package repositories

import (
	"database/sql"

	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/lib/pq"
)

func InsertEvent(db *sql.DB, event models.Event) error {
	query := `
        INSERT INTO Events (
            id, title, room, tags, startTime, endTime, type, location, date, githubRepo, 
            slidesURL, imageSrc, virtualURL, extraImageSrcs, description, about, 
            attendeeIds, organizerIds, usersAttendedIds, createdAt, updatedAt, createdBy
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
            $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
        )
    `

	_, err := db.Exec(query,
		event.ID, event.Title, event.Room, pq.Array(event.Tags), event.StartTime, event.EndTime,
		event.Type, event.Location, event.Date, event.GithubRepo, event.SlidesURL, event.ImageSrc,
		event.VirtualURL, pq.Array(event.ExtraImageSrcs), event.Description, event.About,
		pq.Array(event.AttendeeIds), pq.Array(event.OrganizerIds), pq.Array(event.UsersAttendedIds),
		event.CreatedAt, event.UpdatedAt, event.CreatedBy,
	)

	return err
}
