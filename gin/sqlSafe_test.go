package gin

import (
	"reflect"
	"testing"
)

// RuleSql sql规则
type RuleSql struct {
	ClassType  string   `json:"classType"`
	Group      []string `json:"group"`
	Select     []string `json:"select"`
	Where      string   `json:"where"`
	OrderBy    string   `json:"orderBy"`
	GapType    string   `json:"gapType"`
	FirstYear  int      `json:"firstYear"`
	SecondYear int      `json:"secondYear"`
	StartGap   string   `json:"startGap"`
	EndGap     string   `json:"endGap"`
	Dsn        string   `json:"dsn"`
}

func TestSqlSafe(t *testing.T) {
	// Example data
	ruleSql := RuleSql{
		ClassType:  "example",
		Group:      []string{"group1", "group\n2"},
		Select:     []string{"field1", "field2", "--"},
		Where:      "condition",
		OrderBy:    "order",
		GapType:    "type",
		FirstYear:  2000,
		SecondYear: 2020,
		StartGap:   "start",
		EndGap:     "end",
		Dsn:        "dsn",
	}
	// Start the recursive field checking

	err := ValidateSql(reflect.ValueOf(ruleSql), "ruleSql")
	if err != nil {
		t.Logf("validateRuleSql failed: %#v", err)
	}
	EscapeFields(reflect.ValueOf(&ruleSql), "ruleSql")
	t.Log(ruleSql)
}
