package models

type RoomType int

const (
	Lecture    RoomType = iota // 0
	Classroom                  // 1
	Auditorium                 // 2
	Other                      // 3
)

var RoomTypeNames = map[RoomType]string{
	Lecture:    "Lecture",
	Classroom:  "Classroom",
	Auditorium: "Auditorium",
	Other:      "Other",
}

func (r RoomType) String() string {
	return RoomTypeNames[r]
}
