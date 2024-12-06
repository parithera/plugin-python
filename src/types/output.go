package types

import (
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	"github.com/CodeClarityCE/utility-types/exceptions"
)

type Output struct {
	Result       Result       `json:"result"`
	AnalysisInfo AnalysisInfo `json:"analysis_info"`
}

type Result struct {
	Data any `json:"data"`
}

type AnalysisInfo struct {
	Time   Time                       `json:"time"`
	Errors []exceptions.Error         `json:"errors"`
	Status codeclarity.AnalysisStatus `json:"status"`
	Extra  Extra                      `json:"extra"`
}

type Extra struct {
	VersionSeperator    string `json:"version_seperator"`
	ImportPathSeperator string `json:"import_path_seperator"`
	LockFileVersion     int    `json:"lock_file_version"`
}

type Time struct {
	AnalysisStartTime string  `json:"analysis_start_time"`
	AnalysisEndTime   string  `json:"analysis_end_time"`
	AnalysisDeltaTime float64 `json:"analysis_delta_time"`
}
