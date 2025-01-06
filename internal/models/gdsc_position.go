package models

type GDSCPosition int

const (
	Student GDSCPosition = iota + 1
	Alumni
	Mentor
	Leader
	Advisor
	Sponsor
)

var GDSCPositionMap = map[GDSCPosition]string{
	Student: "Student",
	Alumni:  "Alumni",
	Mentor:  "Mentor",
	Leader:  "Leader",
	Advisor: "Advisor",
	Sponsor: "Sponsor",
}

func (p GDSCPosition) String() string {
	return GDSCPositionMap[p]
}
