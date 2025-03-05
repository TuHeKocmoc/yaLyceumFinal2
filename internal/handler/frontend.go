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

	exprs, err := repository.GetAllExpressions()
	if err != nil {
		http.Error(w, "failed to get expressions", http.StatusInternalServerError)
		return
	}

	data := struct {
		Expressions interface{}
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

	expr := r.FormValue("expression")
	if expr == "" {
		http.Error(w, "empty expression", http.StatusBadRequest)
		return
	}

	if !calc.CheckInput(expr) {
		http.Error(w, "expression is not valid", http.StatusUnprocessableEntity)
		return
	}

	newExpr, err := repository.CreateExpression(expr)
	if err != nil {
		http.Error(w, "cannot create expression", http.StatusInternalServerError)
		return
	}

	_, err = planner.PlanTasks(newExpr.ID, expr)
	if err != nil {
		http.Error(w, "cannot plan tasks: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	newExpr.Status = model.StatusInProgress
	_ = repository.UpdateExpression(newExpr)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleFrontExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL: /expression/<id>
	id := r.URL.Path[len("/expression/"):]

	expr, err := repository.GetExpressionByID(id)
	if err != nil {
		http.Error(w, "repo error", http.StatusInternalServerError)
		return
	}
	if expr == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	data := struct {
		Expression interface{}
	}{
		Expression: expr,
	}

	err = tmplExpression.Execute(w, data)
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
