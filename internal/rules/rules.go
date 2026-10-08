// Package rules provides safe evaluation of expr-lang eligibility
// expressions plus AST utilities used to detect missing profile fields.
package rules

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/expr-lang/expr/vm"
)

// KnownVariables is the set of identifiers that eligibility expressions may
// reference. It includes profile facts and derived fields.
var KnownVariables = map[string]bool{
	"age": true, "gender": true, "marital_status": true, "income_annual": true,
	"occupation": true, "caste_category": true, "residence_type": true,
	"state_code": true, "land_acres": true, "ration_card_type": true,
	"has_bank_account": true, "has_aadhaar": true, "education_level": true,
	"disability_status": true, "minority_status": true,
	"has_pucca_house": true, "has_electricity": true, "employment_status": true,
	"is_pregnant_or_lactating": true, "has_girl_child": true, "girl_child_age": true,
	// Derived
	"is_widow": true, "has_student_child": true, "student_child_class": true,
}

// ValidateVariables ensures every identifier in an expression is known.
func ValidateVariables(expression string) error {
	fields, err := ExtractFields(expression)
	if err != nil {
		return err
	}
	for _, f := range fields {
		if !KnownVariables[f] {
			return fmt.Errorf("unknown field %q in expression", f)
		}
	}
	return nil
}

// Validate compiles an expression as a boolean and reports syntax errors.
func Validate(expression string) error {
	_, err := compile(expression)
	if err != nil {
		return fmt.Errorf("invalid expression: %w", err)
	}
	return nil
}

// Evaluate runs an expression against an environment map and returns a bool.
func Evaluate(expression string, env map[string]any) (bool, error) {
	program, err := compile(expression)
	if err != nil {
		return false, err
	}
	out, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}
	b, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("expression did not evaluate to a boolean")
	}
	return b, nil
}

// ExtractFields returns the identifiers referenced by an expression.
func ExtractFields(expression string) ([]string, error) {
	tree, err := parser.Parse(expression)
	if err != nil {
		return nil, err
	}
	v := &identVisitor{seen: map[string]bool{}}
	node := tree.Node
	ast.Walk(&node, v)
	return v.fields, nil
}

func compile(expression string) (*vm.Program, error) {
	return expr.Compile(expression, expr.AsBool())
}

type identVisitor struct {
	seen   map[string]bool
	fields []string
}

func (v *identVisitor) Visit(node *ast.Node) {
	if id, ok := (*node).(*ast.IdentifierNode); ok {
		if !v.seen[id.Value] {
			v.seen[id.Value] = true
			v.fields = append(v.fields, id.Value)
		}
	}
}
