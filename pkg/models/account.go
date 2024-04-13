package models

// Account представляет банковский счет
type Account struct {
    ID      int     `json:"id"`
    Balance float64 `json:"sum"`
}
