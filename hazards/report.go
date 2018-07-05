package hazards

import (
	"container/list"
	"time"
)

//HazardList a list of hazards
type HazardList = list.List

//HazardReport collection of hazards found during an analysis for a job
type HazardReport struct {
	JobID     string
	CreatedOn time.Time
	Hazards   HazardList
}

//Hazard describes hazardous content that is found in a video
type Hazard struct {
	Start, End uint
	HazardType string
}
