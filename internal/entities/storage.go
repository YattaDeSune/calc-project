package entities

import "time"

type Task struct {
	ID          string `json:"id"`
	Arg1        string `json:"arg1"`
	Arg2        string `json:"arg2"`
	Operation   string `json:"operation"`
	Status      string `json:"status"` // 1.accepted | 2.in progress | 3.completed/error
	Result      any    `json:"result"`
	LastUpdated time.Time
}

type Expression struct {
	ID         int    `json:"id"`
	Expression string `json:"expression"`
	Status     string `json:"status"` // 1.accepted | 2.in progress | 3.completed/error
	Result     any    `json:"result"`
	RPN        []string
	Stack      []string
	Tasks      []*Task
}

// statuses
var (
	Accepted           = "accepted"             // 1
	InProgress         = "in progress"          // 2
	Completed          = "completed"            // 3
	CompletedWithError = "completed with error" // 3
)
