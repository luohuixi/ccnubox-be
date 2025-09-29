package DO

type Seat struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	LabName  string `gorm:"size:100;not null" json:"lab_name"`
	RoomID   string `gorm:"size:100;not null" json:"kind_id"`
	RoomName string `gorm:"size:150;not null" json:"kind_name"`
	DevID    string `gorm:"size:50;not null;uniqueIndex" json:"dev_id"`
	DevName  string `gorm:"size:50;not null" json:"dev_name"`
	Status   string `json:"status"`
}

type TimeSlot struct {
	ID    uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	DevID string `gorm:"index;not null" json:"seat_id"`
	Start string `gorm:"not null" json:"start"`
	End   string `gorm:"not null" json:"end"`
}
