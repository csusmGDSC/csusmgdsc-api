package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID               uuid.UUID   `json:"id,omitempty"`
	Title            string      `json:"title"`
	Room             *CSUSMRoom  `json:"room,omitempty"`
	Tags             []string    `json:"tags,omitempty"`
	StartTime        time.Time   `json:"startTime"`
	EndTime          time.Time   `json:"endTime"`
	Type             int         `json:"type"`
	Location         *string     `json:"location,omitempty"`
	Date             time.Time   `json:"date"`
	GithubRepo       *string     `json:"githubRepo,omitempty"`
	SlidesURL        *string     `json:"slidesURL,omitempty"`
	ImageSrc         *[]byte     `json:"imageSrc,omitempty"`
	VirtualURL       *string     `json:"virtualURL,omitempty"`
	ExtraImageSrcs   [][]byte    `json:"extraImageSrcs,omitempty"`
	Description      string      `json:"description"`
	About            *string     `json:"about,omitempty"`
	AttendeeIds      []uuid.UUID `json:"attendeeIds,omitempty"`
	OrganizerIds     []uuid.UUID `json:"organizerIds"`
	UsersAttendedIds []uuid.UUID `json:"usersAttendedIds,omitempty"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
	CreatedBy        *uuid.UUID  `json:"createdBy,omitempty"`
}

type CSUSMRoom struct {
	Building string `json:"building"`
	Room     int    `json:"room"`
	Type     int    `json:"type"`
	Capacity int    `json:"capacity"`
}
