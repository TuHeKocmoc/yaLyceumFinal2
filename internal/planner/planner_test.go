package planner_test

import (
	"testing"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/planner"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

func TestPlanner_Simple(t *testing.T) {
	expr, err := repository.CreateExpression("2+2*2")
	if err != nil {
		t.Fatalf("CreateExpression failed: %v", err)
	}

	finalTaskID, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err != nil {
		t.Fatalf("PlanTasksWithNestedParen error: %v", err)
	}

	tasks := expr.Tasks
	if len(tasks) == 0 {
		t.Errorf("expected some tasks, got none")
	}

	if finalTaskID <= 0 {
		t.Errorf("unexpected finalTaskID: %d", finalTaskID)
	}
}

func TestPlanner_ExpressionWithParen(t *testing.T) {
	expr, _ := repository.CreateExpression("(2+3)*4")
	finalTaskID, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if finalTaskID <= 0 {
		t.Errorf("finalTaskID is 0")
	}
}

func TestPlanner_InvalidExpression(t *testing.T) {
	expr, _ := repository.CreateExpression("123+")
	_, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
	if err == nil {
		t.Fatal("expected error for incomplete expression, got nil")
	}
}
