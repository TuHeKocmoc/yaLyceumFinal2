package planner_test

import (
	"os"
	"testing"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/planner"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

const testUserID int64 = 1001

func TestMain(m *testing.M) {
	os.Setenv("DB_PATH", ":memory:")
	if err := db.InitDB(); err != nil {
		panic("failed to init db in memory: " + err.Error())
	}
	code := m.Run()
	os.Exit(code)
}

func TestPlanner_Simple(t *testing.T) {
	expr, err := repository.CreateExpression("2+2*2", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression failed: %v", err)
	}

	finalTaskID, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err != nil {
		t.Fatalf("PlanTasksWithNestedParen error: %v", err)
	}

	if finalTaskID <= 0 {
		t.Errorf("unexpected finalTaskID: %d", finalTaskID)
	}

	tasks, err := repository.GetTasksByExpressionID(expr.ID)
	if err != nil {
		t.Fatalf("GetTasksByExpressionID error: %v", err)
	}
	if len(tasks) == 0 {
		t.Errorf("expected some tasks, got none")
	}
}

func TestPlanner_ExpressionWithParen(t *testing.T) {
	expr, err := repository.CreateExpression("(2+3)*4", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	finalTaskID, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if finalTaskID <= 0 {
		t.Errorf("finalTaskID is 0")
	}

	tasks, err := repository.GetTasksByExpressionID(expr.ID)
	if err != nil {
		t.Fatalf("GetTasksByExpressionID error: %v", err)
	}
	if len(tasks) < 2 {
		t.Errorf("expected at least 2 tasks for '(2+3)*4', got %d", len(tasks))
	}
}

func TestPlanner_InvalidExpression(t *testing.T) {
	expr, err := repository.CreateExpression("123+", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	_, err = planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err == nil {
		t.Fatal("expected error for incomplete expression, got nil")
	}
}

func TestPlanTasks_Simple(t *testing.T) {
	repository.Reset()

	expr, err := repository.CreateExpression("2+2*2", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}
	if expr.ID == "" {
		t.Fatal("expr.ID is empty")
	}

	finalID, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err != nil {
		t.Fatalf("PlanTasksWithNestedParen error: %v", err)
	}
	if finalID <= 0 {
		t.Errorf("invalid final task ID = %d, want > 0", finalID)
	}

	err = repository.FetchTasksForExpression(expr)
	if err != nil {
		t.Fatalf("FetchTasksForExpression error: %v", err)
	}
	if len(expr.Tasks) == 0 {
		t.Errorf("expected some tasks, got 0 for expr=%q", expr.Raw)
	}
	t.Logf("Planned tasks: %v", expr.Tasks)
}

func TestPlanTasks_Paren(t *testing.T) {
	repository.Reset()

	expr, err := repository.CreateExpression("(2+3)*4", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	finalID, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if finalID <= 0 {
		t.Errorf("finalTaskID <= 0")
	}
	err = repository.FetchTasksForExpression(expr)
	if err != nil {
		t.Fatalf("FetchTasksForExpression error: %v", err)
	}
	if len(expr.Tasks) < 2 {
		t.Errorf("expected at least 2 tasks for %q, got %d", expr.Raw, len(expr.Tasks))
	}
}

func TestPlanTasks_Invalid(t *testing.T) {
	repository.Reset()

	expr, err := repository.CreateExpression("5/*2", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	_, err = planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err == nil {
		t.Fatal("expected error for invalid expression '5/*2', got nil")
	}
	t.Logf("Got expected error: %v", err)
}
