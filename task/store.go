package task

import (
	"fmt"
	"strings"
)

type Store struct {
	tasks     map[int]*Task 
	nextID    int           
	processor Processor
}

func NewStore(p Processor) *Store {
	return &Store{
		tasks:     make(map[int]*Task),
		nextID:    1,
		processor: p,
	}
}

func validate(text string) error {
	if strings.TrimSpace(text) == "" {
		return ErrEmptyInput
	}
	return nil
}

func (s *Store) Create(text string) (*Task, error) {
	if err := validate(text); err != nil {
		return nil, err
	}
	t := &Task{ID: s.nextID, Text: text, Status: Pending}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}

func (s *Store) Get(id int) (*Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("%w: id=%d", ErrTaskNotFound, id)
	}
	return t, nil
}

func (s *Store) Process(id int) error {
	t, err := s.Get(id)
	if err != nil {
		return err
	}
	if err := t.transition(Processing); err != nil {
		return err
	}

	res, err := s.processor.Process(t.Text)
	if err != nil {
		_ = t.transition(Failed) 
		return fmt.Errorf("processing task %d: %w", id, err)
	}

	t.Result = &res
	return t.transition(Completed)
}
