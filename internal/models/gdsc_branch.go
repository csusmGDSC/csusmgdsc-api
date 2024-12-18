package models

type GDSCBranch int

const (
	Projects  GDSCBranch = iota // 0
	Interview                   // 1
	Marketing                   // 2
)

var GDSCBranchMap = map[GDSCBranch]string{
	Projects:  "Projects",
	Interview: "Interview",
	Marketing: "Marketing",
}

func (b GDSCBranch) String() string {
	return GDSCBranchMap[b]
}
