package main

import (
	"fmt"

	"sis/task"
)

func main() {
	store := task.NewStore(task.TextAnalyzer{})

	t, err := store.Create("Hello Go world")
	if err != nil {
		fmt.Println("create error:", err)
		return
	}
	fmt.Printf("created: id=%d status=%s\n", t.ID, t.Status)

	if err := store.Process(t.ID); err != nil {
		fmt.Println("process error:", err)
		return
	}

	done, _ := store.Get(t.ID)
	fmt.Printf("done: status=%s words=%d chars=%d longest=%q\n",
		done.Status, done.Result.Words, done.Result.Characters, done.Result.LongestWord)

	if _, err := store.Create("   "); err != nil {
		fmt.Println("expected error:", err)
	}
}
