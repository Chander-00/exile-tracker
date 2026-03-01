package models

type CreateAccountInput struct {
	AccountName string  `json:"account_name" validate:"required"`
	Player      *string `json:"player"`
}

type UpdateAccountInput struct {
	AccountName string  `json:"account_name" validate:"required"`
	Player      *string `json:"player"`
}
