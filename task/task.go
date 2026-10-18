package task

import (
	"errors"
	"fmt"
)

type Status string

const (
	Pending    Status = "pending"    
	Processing Status = "processing" 
	Completed  Status = "completed"  
	Failed     Status = "failed"     
)

var (
	ErrEmptyInput        = errors.New("input text is empty or whitespace-only")
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidTransition = errors.New("invalid status transition")
)

var allowedTransitions = map[Status][]Status{
	Pending:    {Processing},
	Processing: {Completed, Failed},
	Completed:  {},
	Failed:     {}, 
}

type Result struct {
	Words       int    
	Characters  int    
	LongestWord string 
}

type Task struct {
	ID     int
	Text   string  
	Status Status  
	Result *Result 
}


func (t *Task) transition(to Status) error {
	for _, allowed := range allowedTransitions[t.Status] {
		if allowed == to {
			t.Status = to
			return nil
		}
	}
	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, t.Status, to)
}
