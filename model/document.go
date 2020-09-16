package model

import (
	"database/sql/driver"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pborman/uuid"
	"gorm.io/gorm"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"
)

type JSONB map[string]interface{}

type Document struct {
	ID        uuid.UUID `gorm:"primary_key;auto_increment" json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Data      []Details `json:"data"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Details struct {
	ID         uint      `gorm:"primary_key;auto_increment" json:"id"`
	DocumentID uuid.UUID `gorm:"not null" json:"document_id"`
	Data       JSONB     `type:jsonb not null default '{}'::jsonb`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (j JSONB) Value() (driver.Value, error) {
	b, err := json.Marshal(j)
	return b, err
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	i := JSONB{}
	err := json.Unmarshal(bytes, &i)

	if err != nil {
		return err
	}

	*j = i

	return nil
}

func (d *Document) PrepareDocument(fname string) {
	d.ID = uuid.NewRandom()
	d.UserID = uuid.NewRandom()
	d.Title = strings.TrimSpace(fname)
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
}

func (d *Details) PrepareDetails(uid uuid.UUID, data JSONB) {
	d.DocumentID = uid
	d.Data = data
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
}

func (d *Document) CreateDocument(file multipart.File, fname string, db *gorm.DB) (*Document, error) {
	var err error

	d.PrepareDocument(fname)

	err = db.Create(&d).Error

	if err != nil {
		return &Document{}, err
	}

	err = CSV2Map(file, d.ID, db)

	if err != nil {
		return &Document{}, err
	}

	details := []Details{}

	err = db.Model(&Details{}).Where("document_id = ?", d.ID).Find(&details).Error

	if err != nil {
		return &Document{}, err
	}

	d.Data = details

	return d, nil
}

func CSV2Map(file multipart.File, uid uuid.UUID, db *gorm.DB) error {
	r := csv.NewReader(file)

	var header []string

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if header == nil {
			header = record
		} else {
			dict := JSONB{}
			for i := range header {
				dict[header[i]] = record[i]
			}

			val, err := dict.Value()
			dict.Scan(val)

			details := Details{}
			if err != nil {
				return err
			}
			details.PrepareDetails(uid, dict)

			err = db.Create(&details).Error

			if err != nil {
				return err
			}
		}
	}

	return nil
}
