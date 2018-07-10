package hazards

import (
	"bytes"
	"crypto/tls"
	"errors"
	"net/http"
)

const reportServerUri = "https://localhost:5001/hazards"

//UploadHazardReport sends the hazard report to the server
func UploadHazardReport(report HazardReport) error {
	json, err := report.MarshalJSON()

	if err != nil {
		return err
	}

	reader := bytes.NewReader(json)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Post(reportServerUri, "application/json", reader)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Server returned non successful status " + resp.Status)
	}

	return nil
}
