package store

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestProjectTodoFlow(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := s.CreateUser(ctx, "USER@example.com", "User", string(hash))
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	project, err := s.CreateProject(ctx, user.ID, "Launch", "Ship todo app", "#22c55e")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	first, err := s.CreateTodo(ctx, user.ID, project.ID, "One", "", "high", nil)
	if err != nil {
		t.Fatalf("create first todo: %v", err)
	}
	second, err := s.CreateTodo(ctx, user.ID, project.ID, "Two", "", "normal", nil)
	if err != nil {
		t.Fatalf("create second todo: %v", err)
	}
	if first.Position != 1 || second.Position != 1 {
		t.Fatalf("unexpected positions: %d %d", first.Position, second.Position)
	}
	todos, err := s.ListTodos(ctx, project.ID)
	if err != nil {
		t.Fatalf("list todos after create: %v", err)
	}
	if len(todos) != 2 || todos[0].ID != second.ID || todos[1].ID != first.ID {
		t.Fatalf("new todo should be listed first: %#v", todos)
	}
	if err := s.ReorderTodos(ctx, project.ID, []string{second.ID, first.ID}); err != nil {
		t.Fatalf("reorder todos: %v", err)
	}
	todos, err = s.ListTodos(ctx, project.ID)
	if err != nil {
		t.Fatalf("list todos: %v", err)
	}
	if len(todos) != 2 || todos[0].ID != second.ID || todos[1].ID != first.ID {
		t.Fatalf("unexpected order: %#v", todos)
	}
	completed := true
	updated, err := s.UpdateTodo(ctx, first.ID, &completed, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("update todo: %v", err)
	}
	if !updated.Completed {
		t.Fatal("todo should be completed")
	}
}
