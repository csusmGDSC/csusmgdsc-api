package models

type GDSCBranch int

const (
	Projects GDSCBranch = iota
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
