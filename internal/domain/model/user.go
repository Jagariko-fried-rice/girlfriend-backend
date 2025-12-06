package model

import "github.com/google/uuid"

// バッチ処理用にユーザーとパートナー情報をまとめて扱う構造体
type UserWithPartner struct {
	UserID       uuid.UUID
	PartnerID    uuid.UUID
	PartnerName  string
	CurrentStage string
	Stamina      int
	Intelligence int
	Sense        int

	Personality  string
	HairColor    string
}