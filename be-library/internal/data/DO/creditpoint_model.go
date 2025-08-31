package DO

type CreditPoint struct {
	StuID  string `gorm:"primaryKey;size:20;not null"`
	System string `gorm:"size:100;not null"`
	Remain string `gorm:"size:100;not null"`
	Total  string `gorm:"size:100;not null"`
}

type CreditRecord struct {
	StuID    string `gorm:"primaryKey;size:20;not null"`
	Title    string `gorm:"size:100;not null"`
	Subtitle string `gorm:"size:150;not null"`
	Location string `gorm:"size:100;not null"`
}
