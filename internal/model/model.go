package model

// "Expression"
const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusDone       = "DONE"
	StatusError      = "ERROR"
)

// "Task"
const (
	TaskStatusWaiting    = "WAITING"
	TaskStatusInProgress = "IN_PROGRESS"
	TaskStatusDone       = "DONE"
	TaskStatusError      = "ERROR"
)

type Expression struct {
	ID     string   `json:"id"`
	Raw    string   `json:"raw"`
	Status string   `json:"status"`
	Result *float64 `json:"result"`
	Tasks  []int    `json:"tasks"`
}

type Task struct {
	ID           int    `json:"id"`
	ExpressionID string `json:"expression_id"`

	Arg1   *float64 `json:"arg1,omitempty"`
	Arg2   *float64 `json:"arg2,omitempty"`
	Op     string   `json:"op"`
	Result *float64 `json:"result"`

	Status string `json:"status"`
}
