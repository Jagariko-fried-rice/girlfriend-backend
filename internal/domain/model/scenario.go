package model

import "github.com/google/uuid"

// Scenario は scenarios テーブルの1行に対応する構造体です
type Scenario struct {
	ID             uuid.UUID
	Stage          string
	Routes         string
	TemplateText   string
	StatEffect     string // 一旦JSON文字列として扱います
	Weight         int
	ConditionStat  *string // NULL許容
	ConditionValue int
	SuccessText    *string // NULL許容
	FailureText    *string // NULL許容
	ImagePrompt    string 
}