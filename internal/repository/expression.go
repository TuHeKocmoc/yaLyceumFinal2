package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
)

func CreateExpression(raw string, userID int64) (*model.Expression, error) {
	exprID := uuid.New().String()
	status := model.StatusPending

	query := `
        INSERT INTO expressions (id, user_id, raw, status)
        VALUES (?, ?, ?, ?)
    `
	_, err := db.GlobalDB.Exec(query, exprID, userID, raw, status)
	if err != nil {
		return nil, fmt.Errorf("create expression error: %w", err)
	}

	return &model.Expression{
		ID:     exprID,
		UserID: userID,
		Raw:    raw,
		Status: status,
	}, nil
}

func GetExpressionByID(userID int64, exprID string) (*model.Expression, error) {
	query := `
        SELECT id, user_id, raw, status, result, final_task_id
        FROM expressions
        WHERE id = ? AND user_id = ?
        LIMIT 1
    `
	row := db.GlobalDB.QueryRow(query, exprID, userID)

	var e model.Expression
	var nullableRes sql.NullFloat64
	var nullableFinalTaskID sql.NullInt64

	err := row.Scan(
		&e.ID,
		&e.UserID,
		&e.Raw,
		&e.Status,
		&nullableRes,
		&nullableFinalTaskID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get expression error: %w", err)
	}

	if nullableRes.Valid {
		val := nullableRes.Float64
		e.Result = &val
	}
	if nullableFinalTaskID.Valid {
		e.FinalTaskID = int(nullableFinalTaskID.Int64)
	} else {
		e.FinalTaskID = 0
	}

	return &e, nil
}

func GetAllExpressions(userID int64) ([]*model.Expression, error) {
	query := `
        SELECT id, user_id, raw, status, result, final_task_id
        FROM expressions
        WHERE user_id = ?
        ORDER BY id
    `
	rows, err := db.GlobalDB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("get all expressions error: %w", err)
	}
	defer rows.Close()

	var result []*model.Expression
	for rows.Next() {
		var e model.Expression
		var nullableRes sql.NullFloat64
		var nullableFinalTaskID sql.NullInt64

		err := rows.Scan(
			&e.ID,
			&e.UserID,
			&e.Raw,
			&e.Status,
			&nullableRes,
			&nullableFinalTaskID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan expression error: %w", err)
		}

		if nullableRes.Valid {
			val := nullableRes.Float64
			e.Result = &val
		}
		if nullableFinalTaskID.Valid {
			e.FinalTaskID = int(nullableFinalTaskID.Int64)
		} else {
			e.FinalTaskID = 0
		}

		result = append(result, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}

func UpdateExpression(e *model.Expression) error {
	query := `
        UPDATE expressions
        SET status = ?,
            result = ?,
            final_task_id = ?
        WHERE id = ? AND user_id = ?
    `
	var resultVal interface{}
	if e.Result != nil {
		resultVal = *e.Result
	} else {
		resultVal = nil
	}

	_, err := db.GlobalDB.Exec(
		query,
		e.Status,
		resultVal,
		e.FinalTaskID,
		e.ID,
		e.UserID,
	)
	if err != nil {
		return fmt.Errorf("update expression error: %w", err)
	}
	return nil
}

func DeleteExpression(userID int64, exprID string) error {
	query := `DELETE FROM expressions WHERE id = ? AND user_id = ?`
	_, err := db.GlobalDB.Exec(query, exprID, userID)
	return err
}

func FetchTasksForExpression(e *model.Expression) error {
	query := `SELECT id FROM tasks WHERE expression_id = ?`
	rows, err := db.GlobalDB.Query(query, e.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var taskIDs []int
	for rows.Next() {
		var tid int
		if err := rows.Scan(&tid); err != nil {
			return err
		}
		taskIDs = append(taskIDs, tid)
	}
	e.Tasks = taskIDs
	return nil
}

func GetExpressionByIDForTask(exprID string) (*model.Expression, error) {
	query := `
        SELECT id, user_id, raw, status, result, final_task_id
        FROM expressions
        WHERE id = ?
        LIMIT 1
    `
	row := db.GlobalDB.QueryRow(query, exprID)

	var e model.Expression
	var nullableRes sql.NullFloat64
	var nullableFinalTaskID sql.NullInt64

	err := row.Scan(
		&e.ID,
		&e.UserID,
		&e.Raw,
		&e.Status,
		&nullableRes,
		&nullableFinalTaskID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetExpressionByIDForTask error: %w", err)
	}

	if nullableRes.Valid {
		val := nullableRes.Float64
		e.Result = &val
	}
	if nullableFinalTaskID.Valid {
		e.FinalTaskID = int(nullableFinalTaskID.Int64)
	} else {
		e.FinalTaskID = 0
	}
	return &e, nil
}

func GetExpressionByIDNoUserCheck(exprID string) (*model.Expression, error) {
	query := `
        SELECT id, user_id, raw, status, result, final_task_id
        FROM expressions
        WHERE id = ?
        LIMIT 1
    `
	row := db.GlobalDB.QueryRow(query, exprID)

	var e model.Expression
	var nullableRes sql.NullFloat64
	var nullableFinalTaskID sql.NullInt64

	err := row.Scan(
		&e.ID,
		&e.UserID,
		&e.Raw,
		&e.Status,
		&nullableRes,
		&nullableFinalTaskID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if nullableRes.Valid {
		val := nullableRes.Float64
		e.Result = &val
	}

	if nullableFinalTaskID.Valid {
		e.FinalTaskID = int(nullableFinalTaskID.Int64)
	} else {
		e.FinalTaskID = 0
	}
	return &e, nil
}
