package hazards

import (
	"bytes"
	"errors"
	"net/http"
)

const reportServerUri = "http://localhost:5001/hazards"

//UploadHazardReport sends the hazard report to the server
func UploadHazardReport(report HazardReport) error {
	json, err := report.MarshalJSON()

	if err != nil {
		return err
	}

	reader := bytes.NewReader(json)
	resp, err := http.Post(reportServerUri, "application/json", reader)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Server returned non successful status " + resp.Status)
	}

	return nil
}
