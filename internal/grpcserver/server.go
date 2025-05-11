package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	calc "github.com/TuHeKocmoc/yalyceumfinal2/internal/proto"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

type CalcServer struct {
	calc.UnimplementedCalcServiceServer
}

func NewGRPCServer() *grpc.Server {
	s := grpc.NewServer()
	srv := &CalcServer{}
	calc.RegisterCalcServiceServer(s, srv)
	return s
}

func (s *CalcServer) GetTask(ctx context.Context, req *calc.GetTaskRequest) (*calc.GetTaskResponse, error) {
	task, err := repository.GetNextWaitingTask()
	if err != nil {
		log.Printf("GetNextWaitingTask error: %v", err)
		return &calc.GetTaskResponse{Status: "ERROR"}, fmt.Errorf("cannot get next task: %w", err)
	}
	if task == nil {
		return &calc.GetTaskResponse{Status: "NO_TASK"}, nil
	}

	task.Status = model.TaskStatusInProgress
	if err := repository.UpdateTask(task); err != nil {
		log.Printf("UpdateTask error: %v", err)
		return &calc.GetTaskResponse{Status: "ERROR"}, err
	}

	operationTime := getOperationTime(task.Op)

	a, b := fetchTaskArgs(task)

	resp := &calc.GetTaskResponse{
		Status: "OK",
		Task: &calc.TaskData{
			Id:            int32(task.ID),
			Arg1:          a,
			Arg2:          b,
			Operation:     task.Op,
			OperationTime: int32(operationTime),
		},
	}
	return resp, nil
}

func (s *CalcServer) PostResult(ctx context.Context, req *calc.PostResultRequest) (*calc.PostResultResponse, error) {
	taskID := int(req.Id)
	task, err := repository.GetTaskByID(taskID)
	if err != nil {
		log.Printf("GetTaskByID error: %v", err)
		return &calc.PostResultResponse{Status: "ERROR"}, err
	}
	if task == nil {
		return &calc.PostResultResponse{Status: "NOT_FOUND"}, nil
	}
	if task.Status != model.TaskStatusInProgress {
		return &calc.PostResultResponse{Status: "BAD_STATUS"}, nil
	}

	resultVal := float64(req.Result)
	task.Result = &resultVal
	task.Status = model.TaskStatusDone
	if err := repository.UpdateTask(task); err != nil {
		log.Printf("UpdateTask error: %v", err)
		return &calc.PostResultResponse{Status: "ERROR"}, err
	}

	expr, err := repository.GetExpressionByIDForTask(task.ExpressionID)
	if err != nil {
		log.Printf("GetExpressionByIDForTask error: %v", err)
		return &calc.PostResultResponse{Status: "ERROR"}, err
	}
	if expr == nil {
		return &calc.PostResultResponse{Status: "ERROR"}, nil
	}

	if task.Op == "FULL" {
		expr.Result = &resultVal
		expr.Status = model.StatusDone
		_ = repository.UpdateExpression(expr)
	} else {
		allDone, lastRes, err := checkAllTasksDone(expr.ID)
		if err != nil {
			log.Printf("checkAllTasksDone error: %v", err)
			return &calc.PostResultResponse{Status: "ERROR"}, err
		}
		if allDone {
			expr.Status = model.StatusDone
			expr.Result = lastRes
			_ = repository.UpdateExpression(expr)
		}
	}

	return &calc.PostResultResponse{Status: "OK"}, nil
}

func fetchTaskArgs(t *model.Task) (float64, float64) {
	var a, b float64

	if t.Arg1Value != nil {
		a = *t.Arg1Value
	} else if t.Arg1TaskID != nil {
		depTask, _ := repository.GetTaskByID(*t.Arg1TaskID)
		if depTask != nil && depTask.Result != nil {
			a = *depTask.Result
		}
	}

	if t.Arg2Value != nil {
		b = *t.Arg2Value
	} else if t.Arg2TaskID != nil {
		depTask, _ := repository.GetTaskByID(*t.Arg2TaskID)
		if depTask != nil && depTask.Result != nil {
			b = *depTask.Result
		}
	}

	return a, b
}

func checkAllTasksDone(exprID string) (bool, *float64, error) {
	tasks, err := repository.GetTasksByExpressionID(exprID)
	if err != nil {
		return false, nil, err
	}
	if len(tasks) == 0 {
		return true, nil, nil
	}

	var lastRes *float64
	var lastID int
	allDone := true

	for _, t := range tasks {
		if t.Status != model.TaskStatusDone {
			allDone = false
		}
		if t.ID > lastID && t.Result != nil {
			lastRes = t.Result
			lastID = t.ID
		}
	}
	return allDone, lastRes, nil
}

func getOperationTime(op string) int {
	switch op {
	case "+", "ADD":
		return 1000
	case "-", "SUB":
		return 1200
	case "*", "MUL":
		return 2000
	case "/", "DIV":
		return 2500
	case "FULL":
		return 3000
	default:
		return 1000
	}
}

func StartGRPCServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := NewGRPCServer()
	log.Printf("Starting gRPC server on %s...", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
