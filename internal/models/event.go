package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID            uuid.UUID  `json:"id,omitempty"`
	Title         string     `json:"title" validate:"required"`
	Room          *CSUSMRoom `json:"room,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	StartTime     time.Time  `json:"start_time" validate:"required"`
	EndTime       time.Time  `json:"end_time" validate:"required"`
	Type          EventType  `json:"type" validate:"required"`
	Location      *string    `json:"location,omitempty"`
	Date          time.Time  `json:"date" validate:"required"`
	RepositoryURL *string    `json:"repository_url,omitempty"`
	SlidesURL     *string    `json:"slides_url,omitempty"`
	ImageSrc      *string    `json:"image_src,omitempty"`
	VirtualURL    *string    `json:"virtual_url,omitempty"`
	Description   string     `json:"description" validate:"required"`
	About         *string    `json:"about,omitempty"`
	CreatedAt     time.Time  `json:"created_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at,omitempty"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
}

type UpdateEventRequest struct {
	ID            *uuid.UUID `json:"id,omitempty" db:"id"`
	Title         *string    `json:"title,omitempty"`
	Room          *CSUSMRoom `json:"room,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Type          *EventType `json:"type,omitempty"`
	Location      *string    `json:"location,omitempty"`
	Date          *time.Time `json:"date,omitempty"`
	RepositoryURL *string    `json:"repository_url,omitempty"`
	SlidesURL     *string    `json:"slides_url,omitempty"`
	ImageSrc      *string    `json:"image_src,omitempty"`
	VirtualURL    *string    `json:"virtual_url,omitempty"`
	Description   *string    `json:"description,omitempty"`
	About         *string    `json:"about,omitempty"`
}

type EventOrganizer struct {
	EventID   uuid.UUID `json:"event_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type AllEventsResponse struct {
	Events     []*Event `json:"events"`
	TotalCount int      `json:"totalCount"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
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
