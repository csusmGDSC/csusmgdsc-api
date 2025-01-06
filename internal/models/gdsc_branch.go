package models

type GDSCBranch int

// Enum types must be plus 1 since 0 value in JSON is registered as empty
const (
	Projects GDSCBranch = iota + 1
	Interview
	Marketing
)

var GDSCBranchMap = map[GDSCBranch]string{
	Projects:  "Projects",
	Interview: "Interview",
	Marketing: "Marketing",
}

func (b GDSCBranch) String() string {
	return GDSCBranchMap[b]
}
