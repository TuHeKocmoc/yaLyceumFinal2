package model

const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusDone       = "DONE"
	StatusError      = "ERROR"
)

const (
	TaskStatusWaiting    = "WAITING"
	TaskStatusInProgress = "IN_PROGRESS"
	TaskStatusDone       = "DONE"
	TaskStatusError      = "ERROR"
)

type Expression struct {
	ID          string   `json:"id"`
	Raw         string   `json:"raw"`
	Status      string   `json:"status"`
	Result      *float64 `json:"result"`
	Tasks       []int    `json:"tasks"`
	FinalTaskID int      `json:"final_task_id,omitempty"`
	UserID      int64    `json:"user_id"`
}

type Task struct {
	ID           int    `json:"id"`
	ExpressionID string `json:"expression_id"`

	Arg1Value  *float64 `json:"arg1_value,omitempty"`
	Arg1TaskID *int     `json:"arg1_task_id,omitempty"`

	Arg2Value  *float64 `json:"arg2_value,omitempty"`
	Arg2TaskID *int     `json:"arg2_task_id,omitempty"`
	Op         string   `json:"op"`
	Result     *float64 `json:"result"`

	Status string `json:"status"`
}
