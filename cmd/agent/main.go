package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	protocalc "github.com/TuHeKocmoc/yalyceumfinal2/internal/proto"
)

var (
	grpcAddr       = "localhost:50051"
	computingPower = 1
)

func main() {
	cpStr := os.Getenv("COMPUTING_POWER")
	if cpStr == "" {
		cpStr = "1"
	}
	cp, err := strconv.Atoi(cpStr)
	if err != nil {
		log.Printf("Invalid COMPUTING_POWER=%s, use 1 by default\n", cpStr)
		cp = 1
	}
	computingPower = cp

	addrFromEnv := os.Getenv("GRPC_ADDR")
	if addrFromEnv != "" {
		grpcAddr = addrFromEnv
	}

	log.Printf("[AGENT] Starting with %d workers. gRPC server = %s\n",
		computingPower, grpcAddr)

	conn, err := grpc.NewClient(
		"dns:///"+grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create new gRPC client: %v", err)
	}
	defer conn.Close()

	client := protocalc.NewCalcServiceClient(conn)

	for i := 0; i < computingPower; i++ {
		go worker(i, client)
	}

	select {}
}

func worker(workerID int, client protocalc.CalcServiceClient) {
	log.Printf("[Worker #%d] started", workerID)

	for {
		time.Sleep(2 * time.Second)

		gtResp, err := client.GetTask(context.Background(), &protocalc.GetTaskRequest{})
		if err != nil {
			log.Printf("[Worker #%d] GetTask error: %v", workerID, err)
			continue
		}

		if gtResp.Status == "NO_TASK" {
			continue
		}
		if gtResp.Status != "OK" {
			log.Printf("[Worker #%d] unexpected GetTask status: %s", workerID, gtResp.Status)
			continue
		}

		task := gtResp.Task
		log.Printf("[Worker #%d] got task ID=%d, op=%s, arg1=%.2f, arg2=%.2f, opTime=%d",
			workerID,
			task.Id,
			task.Operation,
			task.Arg1,
			task.Arg2,
			task.OperationTime,
		)

		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

		resultValue, err := compute(task.Arg1, task.Arg2, task.Operation)
		if err != nil {
			log.Printf("[Worker #%d] compute error: %v", workerID, err)
			continue
		}

		prReq := &protocalc.PostResultRequest{
			Id:     task.Id,
			Result: resultValue,
		}
		prResp, err := client.PostResult(context.Background(), prReq)
		if err != nil {
			log.Printf("[Worker #%d] PostResult error: %v", workerID, err)
			continue
		}
		if prResp.Status != "OK" {
			log.Printf("[Worker #%d] PostResult status=%s", workerID, prResp.Status)
			continue
		}

		log.Printf("[Worker #%d] done task ID=%d, result=%.2f", workerID, task.Id, resultValue)
	}
}

func compute(a, b float64, op string) (float64, error) {
	switch op {
	case "FULL":
		return 42, nil
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", op)
	}
}
