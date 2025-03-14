package output_generator

import (
	"time"

	sbomTypes "github.com/CodeClarityCE/plugin-sbom-javascript/src/types/sbom/js"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
)

// GetAnalysisTiming determines the analysis start time, end time, and duration.
// It accepts the analysis start time as input and returns the formatted start time, end time, and elapsed time in seconds.
func GetAnalysisTiming(start time.Time) (string, string, float64) {
	end := time.Now()
	elapsed := time.Since(start)
	return start.Local().String(), end.Local().String(), elapsed.Seconds()
}

// WriteFailureOutput constructs a failure output for the analysis process.
// It sets the analysis status to 'FAILURE', records the analysis timing details, retrieves any encountered errors, and returns the updated output.
func WriteFailureOutput(output sbomTypes.Output, start time.Time) sbomTypes.Output {
	output.AnalysisInfo.Status = codeclarity.FAILURE
	formattedStart, formattedEnd, delta := GetAnalysisTiming(start)
	output.AnalysisInfo.Time.AnalysisStartTime = formattedStart
	output.AnalysisInfo.Time.AnalysisEndTime = formattedEnd
	output.AnalysisInfo.Time.AnalysisDeltaTime = delta

	output.AnalysisInfo.Errors = exceptionManager.GetErrors()

	return output
}
