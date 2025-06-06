package handler

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/calc"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/planner"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

var (
	tmplIndex      *template.Template
	tmplExpression *template.Template
)

func InitTemplates() error {
	var err error
	tmplIndex, err = template.ParseFiles(
		filepath.Join("web", "index.html"),
	)
	if err != nil {
		return err
	}
	tmplExpression, err = template.ParseFiles(
		filepath.Join("web", "expression.html"),
	)
	if err != nil {
		return err
	}
	return nil
}

func HandleFrontIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	exprs, err := repository.GetAllExpressions(userID)
	if err != nil {
		http.Error(w, "failed to get expressions", http.StatusInternalServerError)
		return
	}

	data := struct {
		Expressions []*model.Expression
	}{
		Expressions: exprs,
	}

	err = tmplIndex.Execute(w, data)
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

func HandleFrontAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	expr := r.FormValue("expression")
	if expr == "" {
		http.Error(w, "empty expression", http.StatusBadRequest)
		return
	}

	if !calc.CheckInput(expr) {
		http.Error(w, "expression is not valid", http.StatusUnprocessableEntity)
		return
	}

	newExpr, err := repository.CreateExpression(expr, userID)
	if err != nil {
		http.Error(w, "cannot create expression", http.StatusInternalServerError)
		return
	}

	finalTaskID, err := planner.PlanTasksWithNestedParen(newExpr.ID, expr)
	if err != nil {
		newExpr.Status = model.StatusError
		_ = repository.UpdateExpression(newExpr)
		http.Error(w, "cannot plan tasks: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	newExpr.Status = model.StatusInProgress
	newExpr.FinalTaskID = finalTaskID
	_ = repository.UpdateExpression(newExpr)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleFrontExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.URL.Path[len("/expression/"):]

	expr, err := repository.GetExpressionByID(userID, id)
	if err != nil {
		http.Error(w, "repo error", http.StatusInternalServerError)
		return
	}
	if expr == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	data := struct {
		Expression *model.Expression
	}{
		Expression: expr,
	}

	err = tmplExpression.Execute(w, data)
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
