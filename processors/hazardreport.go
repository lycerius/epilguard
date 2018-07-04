package processors

import "time"

//HazardReport collection of hazards found during an analysis for a job
type HazardReport struct {
	JobID   string
	Date    time.Time
	Hazards []Hazard
}

//Hazard describes hazardous content that is found in a video
type Hazard struct {
	Start, End uint
	HazardType string
}
