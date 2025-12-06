package model

import "github.com/google/uuid"

// Scenario は scenarios テーブルの1行に対応する構造体です
type Scenario struct {
	ID             uuid.UUID
	Stage          string
	Routes         string
	TemplateText   string
	StatEffect     string // JSON文字列
	Weight         int
	ConditionStat  *string // NULL許容
	ConditionValue int
	SuccessText    *string // NULL許容
	FailureText    *string // NULL許容
	
	SuccessEffect  string // JSON文字列
	FailureEffect  string // JSON文字列
	ImagePrompt    string 
}