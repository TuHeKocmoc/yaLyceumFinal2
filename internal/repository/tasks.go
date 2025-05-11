package repository

import (
	"database/sql"
	"fmt"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
)

func GetTasksByExpressionID(exprID string) ([]*model.Task, error) {
	query := `
        SELECT id, expression_id, op,
               arg1_value, arg1_task_id,
               arg2_value, arg2_task_id,
               result, status
        FROM tasks
        WHERE expression_id = ?
        ORDER BY id
    `
	rows, err := db.GlobalDB.Query(query, exprID)
	if err != nil {
		return nil, fmt.Errorf("GetTasksByExpressionID query error: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var t model.Task

		var arg1Val sql.NullFloat64
		var arg2Val sql.NullFloat64
		var arg1TaskID sql.NullInt64
		var arg2TaskID sql.NullInt64
		var res sql.NullFloat64

		err := rows.Scan(
			&t.ID,
			&t.ExpressionID,
			&t.Op,
			&arg1Val,
			&arg1TaskID,
			&arg2Val,
			&arg2TaskID,
			&res,
			&t.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("GetTasksByExpressionID scan error: %w", err)
		}

		if arg1Val.Valid {
			f := arg1Val.Float64
			t.Arg1Value = &f
		}
		if arg2Val.Valid {
			f := arg2Val.Float64
			t.Arg2Value = &f
		}
		if arg1TaskID.Valid {
			i := int(arg1TaskID.Int64)
			t.Arg1TaskID = &i
		}
		if arg2TaskID.Valid {
			i := int(arg2TaskID.Int64)
			t.Arg2TaskID = &i
		}
		if res.Valid {
			f := res.Float64
			t.Result = &f
		}

		tasks = append(tasks, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
