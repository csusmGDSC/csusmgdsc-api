package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID               uuid.UUID   `json:"id,omitempty"`
	Title            string      `json:"title" validate:"required"`
	Room             *CSUSMRoom  `json:"room,omitempty"`
	Tags             []string    `json:"tags,omitempty"`
	StartTime        time.Time   `json:"startTime" validate:"required"`
	EndTime          time.Time   `json:"endTime" validate:"required"`
	Type             EventType   `json:"type" validate:"required"`
	Location         *string     `json:"location,omitempty"`
	Date             time.Time   `json:"date" validate:"required"`
	GithubRepo       *string     `json:"githubRepo,omitempty"`
	SlidesURL        *string     `json:"slidesURL,omitempty"`
	ImageSrc         *[]byte     `json:"imageSrc,omitempty"`
	VirtualURL       *string     `json:"virtualURL,omitempty"`
	ExtraImageSrcs   [][]byte    `json:"extraImageSrcs,omitempty"`
	Description      string      `json:"description" validate:"required"`
	About            *string     `json:"about,omitempty"`
	AttendeeIds      []uuid.UUID `json:"attendeeIds,omitempty"`
	OrganizerIds     []uuid.UUID `json:"organizerIds" validate:"required"`
	UsersAttendedIds []uuid.UUID `json:"usersAttendedIds,omitempty"`
	CreatedAt        time.Time   `json:"createdAt,omitempty"`
	UpdatedAt        time.Time   `json:"updatedAt,omitempty"`
	CreatedBy        *uuid.UUID  `json:"createdBy,omitempty"`
}

type CSUSMRoom struct {
	Building string   `json:"building"`
	Room     int      `json:"room"`
	Type     RoomType `json:"type"`
	Capacity int      `json:"capacity"`
}

type EventType int

const (
	Virtual EventType = iota + 1
	Leetcode
	Hackathon
	Meeting
	Project
	Workshop
	Competition
	Challenge
)

var EventTypeMap = map[EventType]string{
	Virtual:     "Virtual",
	Leetcode:    "Leetcode",
	Hackathon:   "Hackathon",
	Meeting:     "Meeting",
	Project:     "Project",
	Workshop:    "Workshop",
	Competition: "Competition",
	Challenge:   "Challenge",
}

func (t EventType) String() string {
	return EventTypeMap[t]
}
