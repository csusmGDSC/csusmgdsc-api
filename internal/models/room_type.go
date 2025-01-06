package models

type RoomType int

const (
	Lecture    RoomType = iota + 1 // 1
	Classroom                      // 2
	Auditorium                     // 3
	Other                          // 4
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
