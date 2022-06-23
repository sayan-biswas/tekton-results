package models

import "time"

// Record is the database model of a Record
type Record struct {
	// Result is used to create the relationship between the Result and Records
	// table. Data will not be returned here during reads. Use the foreign key
	// fields instead.
	Result     Result `gorm:"foreignKey:Parent,ResultID;references:Parent,ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Parent     string `gorm:"primaryKey;index:records_by_name,priority:1"`
	ResultID   string `gorm:"primaryKey"`
	ResultName string `gorm:"index:records_by_name,priority:2"`

	ID   string `gorm:"primaryKey"`
	Name string `gorm:"index:records_by_name,priority:3"`

	Type string
	Data []byte

	CreatedTime time.Time
	UpdatedTime time.Time

	Etag string
}
