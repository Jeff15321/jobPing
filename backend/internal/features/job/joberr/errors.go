package joberr

import "errors"

var (
	ErrJobNotFound     = errors.New("job not found")
	ErrAIAnalysisFailed = errors.New("AI analysis failed")
)

