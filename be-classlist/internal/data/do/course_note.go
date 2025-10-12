package do

import "time"

const CourseNoteTableName = "course_note"

type CourseNote struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement"`
    StuID     string    `gorm:"type:varchar(32);not null;index:idx_note_unique,unique"`
    Year      string    `gorm:"type:varchar(8);not null;index:idx_note_unique,unique"`
    Semester  string    `gorm:"type:varchar(4);not null;index:idx_note_unique,unique"`
    ClaID     string    `gorm:"type:varchar(128);not null;index:idx_note_unique,unique"`
    Note      string    `gorm:"type:text;not null"`
    CreatedAt time.Time `gorm:"<-:create"`
    UpdatedAt time.Time
}

func (CourseNote) TableName() string { return CourseNoteTableName }