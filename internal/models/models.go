package models

var DatabaseModels = []interface{}{
	Account{},
}

type Account struct {
	UserID       string `gorm:"primaryKey;type:text;autoIncrement:false" json:"id"`
	DisplayName  string `gorm:"type:text" json:"displayName"`
	Name         string `gorm:"type:text" json:"name"`
	Email        string `gorm:"unique;type:text" json:"email"`
	IsAdmin      bool   `gorm:"default:false" json:"isAdmin"`
	Provider     string `gorm:"type:text" json:"provider"`
	CreatedAt    int64  `json:"createdAt"`
	LastLoggedIn int64  `json:"lastLoggedIn"`
	Picture      string `gorm:"type:text" json:"pictucre"`
}
