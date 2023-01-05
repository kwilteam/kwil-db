package tasks

import (
	"context"
	"database/sql"
)

/*
	Tasks are a set of arbitrary functions that can be executed against a database.
	All tasks will be included in a specific transaction, and are made to handle
	a specified range of blocks and the events that occurred within that range.

	The point of organizing tasks this way is to make it easy to add new smart contract
	events that we want to sync to the database.
*/

type taskRunner struct {
	tasks []Runnable
	final Runnable // the final task to run after all other tasks have completed
}

type Runnable interface {
	Run(ctx context.Context, chunk *Chunk) error
}

type TaskRunner interface {
	Add(task Runnable)
	Runnable
}

type Chunk struct {
	Tx     *sql.Tx
	Start  int64
	Finish int64
}

func New(tasks ...Runnable) *taskRunner {
	return &taskRunner{
		tasks: tasks,
	}
}

func (t *taskRunner) Add(task Runnable) {
	t.tasks = append(t.tasks, task)
}

func (t *taskRunner) Run(ctx context.Context, chunk *Chunk) error {
	for _, task := range t.tasks {
		err := task.Run(ctx, chunk)
		if err != nil {
			return err
		}
	}

	if t.final != nil {
		err := t.final.Run(ctx, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *taskRunner) SetFinal(task Runnable) {
	t.final = task
}
