package contracts

type CheckTask struct {
	TaskID      string `json:"task_id"`
	TargetID    int64  `json:"target_id"`
	URL         string `json:"URL"`
	TimeoutSec  int    `json:"timeout_sec"`
	ScheduledAt string `json:"scheduled_at"`
}

type CheckResult struct {
	TaskID         string `json:"task_id"`
	TargetID       int64  `json:"target_id"`
	Status         string `json:"status"`
	StatusCode     int    `json:"status_code"`
	ErrorMsg       string `json:"error_msg"`
	ResponseTimeMs int    `json:"response_time_ms"`
	CheckedAt      string `json:"checked_at"`
}
