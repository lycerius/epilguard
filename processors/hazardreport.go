package processors

import "time"

type HazardReport struct {
	JobID   string
	Date    time.Time
	Hazards []Hazard
}

type Hazard struct {
	Start, End uint
	HazardType string
}
