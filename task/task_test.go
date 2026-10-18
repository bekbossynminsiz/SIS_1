package task

import (
	"errors"
	"testing"
)

type failingProcessor struct{}

func (failingProcessor) Process(string) (Result, error) {
	return Result{}, errors.New("boom")
}

func newTestStore() *Store { return NewStore(TextAnalyzer{}) }

func TestCreate_Success(t *testing.T) {
	s := newTestStore()
	task, err := s.Create("Hello Go world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 1 {
		t.Errorf("ID = %d, want 1", task.ID)
	}
	if task.Status != Pending {
		t.Errorf("Status = %s, want %s", task.Status, Pending)
	}
	if task.Result != nil {
		t.Errorf("Result should be nil before processing")
	}
}

func TestCreate_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"spaces only", "   "},
		{"tabs and newlines", " \t\n "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore()
			task, err := s.Create(tt.input)
			if !errors.Is(err, ErrEmptyInput) {
				t.Errorf("err = %v, want ErrEmptyInput", err)
			}
			if task != nil {
				t.Errorf("task should be nil on invalid input")
			}
		})
	}
}

func TestGet_Success(t *testing.T) {
	s := newTestStore()
	created, _ := s.Create("some text")
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "some text" {
		t.Errorf("Text = %q, want %q", got.Text, "some text")
	}
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore()
	_, err := s.Get(42)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestProcess_ResultAndState(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantWords   int
		wantChars   int
		wantLongest string
	}{
		{"simple", "Hello Go world", 3, 14, "Hello"},
		{"extra spaces", "  a   bb  ", 2, 10, "bb"},
		{"unicode", "привет мир", 2, 10, "привет"},
		{"single word", "Go", 1, 2, "Go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore()
			created, _ := s.Create(tt.input)

			if err := s.Process(created.ID); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, _ := s.Get(created.ID)
			if got.Status != Completed {
				t.Errorf("Status = %s, want %s", got.Status, Completed)
			}
			if got.Result == nil {
				t.Fatal("Result is nil")
			}
			if got.Result.Words != tt.wantWords {
				t.Errorf("Words = %d, want %d", got.Result.Words, tt.wantWords)
			}
			if got.Result.Characters != tt.wantChars {
				t.Errorf("Characters = %d, want %d", got.Result.Characters, tt.wantChars)
			}
			if got.Result.LongestWord != tt.wantLongest {
				t.Errorf("LongestWord = %q, want %q", got.Result.LongestWord, tt.wantLongest)
			}
		})
	}
}

func TestProcess_NotFound(t *testing.T) {
	s := newTestStore()
	if err := s.Process(99); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestProcess_AlreadyCompleted(t *testing.T) {
	s := newTestStore()
	created, _ := s.Create("hello world")
	if err := s.Process(created.ID); err != nil {
		t.Fatalf("first Process failed: %v", err)
	}

	err := s.Process(created.ID) 
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("err = %v, want ErrInvalidTransition", err)
	}
	got, _ := s.Get(created.ID)
	if got.Status != Completed {
		t.Errorf("Status = %s, want %s (must stay unchanged)", got.Status, Completed)
	}
}

func TestProcess_ProcessorFails(t *testing.T) {
	s := NewStore(failingProcessor{})
	created, _ := s.Create("hello")

	if err := s.Process(created.ID); err == nil {
		t.Fatal("expected error, got nil")
	}
	got, _ := s.Get(created.ID)
	if got.Status != Failed {
		t.Errorf("Status = %s, want %s", got.Status, Failed)
	}
	if got.Result != nil {
		t.Errorf("Result should be nil for failed task")
	}
}

func TestTransition_Table(t *testing.T) {
	tests := []struct {
		from, to Status
		wantErr  bool
	}{
		{Pending, Processing, false},
		{Processing, Completed, false},
		{Processing, Failed, false},
		{Pending, Completed, true},
		{Completed, Processing, true},
		{Failed, Pending, true},
	}
	for _, tt := range tests {
		task := &Task{Status: tt.from}
		err := task.transition(tt.to)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s -> %s: err = %v, wantErr = %v", tt.from, tt.to, err, tt.wantErr)
		}
	}
}

func TestMultipleIndependentTasks(t *testing.T) {
	s := newTestStore()
	a, _ := s.Create("one")
	b, _ := s.Create("two words here")
	c, _ := s.Create("third")

	if a.ID == b.ID || b.ID == c.ID || a.ID == c.ID {
		t.Fatalf("IDs must be unique: %d, %d, %d", a.ID, b.ID, c.ID)
	}

	if err := s.Process(b.ID); err != nil { 
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Status != Pending || c.Status != Pending {
		t.Errorf("other tasks must stay pending: a=%s c=%s", a.Status, c.Status)
	}
	if a.Result != nil || c.Result != nil {
		t.Errorf("other tasks must have no result")
	}
	if b.Result == nil || b.Result.Words != 3 {
		t.Errorf("task b has wrong result: %+v", b.Result)
	}

	if err := s.Process(a.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Result.Words != 3 || a.Result.Words != 1 {
		t.Errorf("results got mixed up: a=%+v b=%+v", a.Result, b.Result)
	}
}
