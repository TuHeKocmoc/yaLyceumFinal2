package repository

import (
	"database/sql"
	"fmt"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
)

func GetTaskByID(id int) (*model.Task, error) {
	query := `
        SELECT
            id,
            expression_id,
            op,
            arg1_value,
            arg1_task_id,
            arg2_value,
            arg2_task_id,
            result,
            status
        FROM tasks
        WHERE id = ?
        LIMIT 1
    `
	row := db.GlobalDB.QueryRow(query, id)

	var t model.Task
	var arg1Val sql.NullFloat64
	var arg1TaskID sql.NullInt64
	var arg2Val sql.NullFloat64
	var arg2TaskID sql.NullInt64
	var resVal sql.NullFloat64

	err := row.Scan(
		&t.ID,
		&t.ExpressionID,
		&t.Op,
		&arg1Val,
		&arg1TaskID,
		&arg2Val,
		&arg2TaskID,
		&resVal,
		&t.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
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
	if resVal.Valid {
		f := resVal.Float64
		t.Result = &f
	}

	return &t, nil
}

func UpdateTask(t *model.Task) error {
	query := `
        UPDATE tasks
        SET
            op = ?,
            arg1_value = ?,
            arg1_task_id = ?,
            arg2_value = ?,
            arg2_task_id = ?,
            result = ?,
            status = ?
        WHERE id = ?
    `
	var (
		arg1Val  interface{}
		arg1Task interface{}
		arg2Val  interface{}
		arg2Task interface{}
		resVal   interface{}
	)

	if t.Arg1Value != nil {
		arg1Val = *t.Arg1Value
	}
	if t.Arg1TaskID != nil {
		arg1Task = *t.Arg1TaskID
	}
	if t.Arg2Value != nil {
		arg2Val = *t.Arg2Value
	}
	if t.Arg2TaskID != nil {
		arg2Task = *t.Arg2TaskID
	}
	if t.Result != nil {
		resVal = *t.Result
	}

	res, err := db.GlobalDB.Exec(query,
		t.Op,
		arg1Val,
		arg1Task,
		arg2Val,
		arg2Task,
		resVal,
		t.Status,
		t.ID,
	)
	if err != nil {
		return fmt.Errorf("UpdateTask exec error: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task not found with id=%d", t.ID)
	}

	return nil
}

func GetNextWaitingTask() (*model.Task, error) {

	waitingQuery := `
        SELECT
            id,
            expression_id,
            op,
            arg1_value,
            arg1_task_id,
            arg2_value,
            arg2_task_id,
            result,
            status
        FROM tasks
        WHERE status = 'WAITING'
        ORDER BY id
    `
	rows, err := db.GlobalDB.Query(waitingQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var tr model.Task
		var arg1Val sql.NullFloat64
		var arg1TaskID sql.NullInt64
		var arg2Val sql.NullFloat64
		var arg2TaskID sql.NullInt64
		var resVal sql.NullFloat64

		err := rows.Scan(
			&tr.ID,
			&tr.ExpressionID,
			&tr.Op,
			&arg1Val,
			&arg1TaskID,
			&arg2Val,
			&arg2TaskID,
			&resVal,
			&tr.Status,
		)
		if err != nil {
			return nil, err
		}
		if arg1Val.Valid {
			f := arg1Val.Float64
			tr.Arg1Value = &f
		}
		if arg2Val.Valid {
			f := arg2Val.Float64
			tr.Arg2Value = &f
		}
		if arg1TaskID.Valid {
			i := int(arg1TaskID.Int64)
			tr.Arg1TaskID = &i
		}
		if arg2TaskID.Valid {
			i := int(arg2TaskID.Int64)
			tr.Arg2TaskID = &i
		}
		if resVal.Valid {
			f := resVal.Float64
			tr.Result = &f
		}
		tasks = append(tasks, &tr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.Arg1TaskID != nil {
			dep, err := GetTaskByID(*t.Arg1TaskID)
			if err != nil {
				return nil, err
			}
			if dep == nil || dep.Status != model.TaskStatusDone {
				continue
			}
		}
		if t.Arg2TaskID != nil {
			dep, err := GetTaskByID(*t.Arg2TaskID)
			if err != nil {
				return nil, err
			}
			if dep == nil || dep.Status != model.TaskStatusDone {
				continue
			}
		}
		return t, nil
	}

	return nil, nil
}

func CreateTaskWithArgs(
	expressionID string,
	op string,
	arg1Value *float64, arg1TaskID *int,
	arg2Value *float64, arg2TaskID *int,
) (*model.Task, error) {

	query := `
        INSERT INTO tasks (
            expression_id,
            op,
            arg1_value,
            arg1_task_id,
            arg2_value,
            arg2_task_id,
            status
        ) VALUES (?, ?, ?, ?, ?, ?, 'WAITING')
    `
	var arg1Val, arg1T, arg2Val, arg2T interface{}
	if arg1Value != nil {
		arg1Val = *arg1Value
	}
	if arg1TaskID != nil {
		arg1T = *arg1TaskID
	}
	if arg2Value != nil {
		arg2Val = *arg2Value
	}
	if arg2TaskID != nil {
		arg2T = *arg2TaskID
	}

	res, err := db.GlobalDB.Exec(query,
		expressionID,
		op,
		arg1Val,
		arg1T,
		arg2Val,
		arg2T,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateTaskWithArgs insert error: %w", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("cannot get lastInsertId: %w", err)
	}

	taskID := int(lastID)
	newTask := &model.Task{
		ID:           taskID,
		ExpressionID: expressionID,
		Op:           op,
		Arg1Value:    arg1Value,
		Arg1TaskID:   arg1TaskID,
		Arg2Value:    arg2Value,
		Arg2TaskID:   arg2TaskID,
		Status:       model.TaskStatusWaiting,
	}

	return newTask, nil
}

func Reset() error {
	_, err := db.GlobalDB.Exec("DELETE FROM tasks;")
	if err != nil {
		return err
	}
	_, err = db.GlobalDB.Exec("DELETE FROM expressions;")
	if err != nil {
		return err
	}
	return nil
}
