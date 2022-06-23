package models

import "time"

// Result is the database model of a Result.
type Result struct {
	Parent      string `gorm:"primaryKey;index:results_by_name,priority:1"`
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"index:results_by_name,priority:2"`
	Annotations Annotations

	CreatedTime time.Time
	UpdatedTime time.Time

	Summary RecordSummary `gorm:"embedded;embeddedPrefix:recordsummary_"`

	Etag string
}

// RecordSummary is the database model of a Result.RecordSummary.
type RecordSummary struct {
	Record      string
	Type        string
	StartTime   *time.Time
	EndTime     *time.Time
	Status      int32
	Annotations Annotations
}
