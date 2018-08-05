package hazards

import (
	"bytes"
	"container/list"
	"encoding/json"
	"time"
)

//HazardList a list of hazards
type HazardList = list.List

//HazardReport collection of hazards found during processing
type HazardReport struct {
	CreatedOn time.Time
	Hazards   HazardList
}

//MarshalJSON converts a hazard report to JSON
func (hr *HazardReport) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteByte('{')

	m, err := json.Marshal(hr.CreatedOn)

	if err != nil {
		return nil, err
	}

	buf.WriteString("\"createdOn\":")
	buf.Write(m)
	buf.WriteByte(',')
	buf.WriteString("\"hazards\":[")
	once := false
	for ele := hr.Hazards.Front(); ele != nil; ele = ele.Next() {
		val := ele.Value

		m, err = json.Marshal(val)

		if err != nil {
			return nil, err
		}

		buf.Write(m)
		buf.WriteByte(',')
		once = true
	}
	if once {
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteByte(']')
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

//Hazard describes hazardous content that is found in a video
type Hazard struct {
	Start      uint   `json:"start"`
	End        uint   `json:"end"`
	HazardType string `json:"hazardType"`
}
