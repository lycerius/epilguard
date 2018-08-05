package test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/lycerius/epilguard/hazards"

	"github.com/lycerius/epilguard/processors"
	"github.com/stretchr/testify/assert"
)

func createTestProcessor(path string, t *assert.Assertions) processors.FlashingProcessor {
	decoder := createDecoderTestDecoder(path, t)
	proc := processors.NewFlashingProcessor(&decoder, Test_Report_Directory)
	createTestDirectory(t)
	return proc
}

func createTestDirectory(t *assert.Assertions) {
	err := os.Mkdir(Test_Report_Directory, 0777)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "file exists") {
			t.NoError(err, "Error creating testing directory")
		}
	}
}

func emptyTestDirectory(t *assert.Assertions) {
	err := os.RemoveAll(Test_Report_Directory)
	t.NoError(err, "Error removing testing directory")
}

func TestCanCreateProcessor(t *testing.T) {
	assert := assert.New(t)
	createTestProcessor(Test_Video_White, assert)
	defer emptyTestDirectory(assert)
}

func TestProcessorCreatesReports(t *testing.T) {
	assert := assert.New(t)
	proc := createTestProcessor(Test_Video_White, assert)
	defer emptyTestDirectory(assert)
	err := proc.Process()
	assert.NoError(err)
	files, err := ioutil.ReadDir(Test_Report_Directory)
	assert.NoError(err)
	assert.Equalf(4, len(files), "Expected 4 report files, got %d", len(files))
}

func TestProcessorDoesntFailWhite(t *testing.T) {
	assert := assert.New(t)
	proc := createTestProcessor(Test_Video_White, assert)
	defer emptyTestDirectory(assert)
	err := proc.Process()
	assert.NoError(err)

	report := proc.HazardReport
	assert.Equalf(0, report.Hazards.Len(), "Expected no hazards, got %d", report.Hazards.Len())
}
func TestProcessorFailsOhhGod(t *testing.T) {
	assert := assert.New(t)
	proc := createTestProcessor(Test_Video_Fail, assert)
	defer emptyTestDirectory(assert)

	err := proc.Process()
	assert.NoError(err)

	report := proc.HazardReport

	assert.Equal(1, report.Hazards.Len(), "Expected 1 hazard, got %d", report.Hazards.Len())

	hazard := report.Hazards.Front().Value.(hazards.Hazard)
	assert.Equal(0, int(hazard.Start), "Expected hazard start to be 0")
	assert.Equal(5, int(hazard.End), "Expected hazard to end at 5")
}
