package entities

type Task struct {
	ID        string `json:"id"`
	Arg1      string `json:"arg1"`
	Arg2      string `json:"arg2"`
	Operation string `json:"operation"`
	Status    string `json:"status"` // accepted | in progress | completed | error
	Result    any    `json:"result"`
}

type Expression struct {
	ID         int    `json:"id"`
	Expression string `json:"expression"`
	Status     string `json:"status"` // accepted | in progress | completed | error
	Result     any    `json:"result"`
	RPN        []string
	Stack      []string
	Tasks      []*Task // вынести отдельно от выражений
}

// statuses
var (
	Accepted           = "accepted"
	InProgress         = "in progress"
	Completed          = "completed"
	CompletedWithError = "completed with error"
)
