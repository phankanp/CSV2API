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
	ID        uuid.UUID `gorm:"primary_key;" json:"id"`
	UserID    uuid.UUID `json:"-"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Header    []Header  `gorm:"not null" json:"-"`
	Data      []Details `gorm:"OnDelete:SET NULL;" json:"data"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
}

type Details struct {
	ID         uint      `gorm:"primary_key;auto_increment" json:"id"`
	DocumentID uuid.UUID `gorm:"not null" json:"-"`
	Data       JSONB     `type:jsonb not null default '{}'::jsonb json:"data"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
}

type Header struct {
	ID         uint      `gorm:"primary_key;auto_increment" json:"id"`
	DocumentID uuid.UUID `gorm:"not null" json:"document_id"`
	Name       string    `gorm:"not null" json:"name"`
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

func (d *Document) PrepareDocument(fname string, uid uuid.UUID) {
	d.ID = uuid.NewRandom()
	d.UserID = uid
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

func (h *Header) PrepareHeader(docID uuid.UUID, name string) {
	h.DocumentID = docID
	h.Name = name
}

func (d *Document) CreateDocument(file multipart.File, fname string, db *gorm.DB, authenticatedUser *User) (*Document, error) {
	var err error

	d.PrepareDocument(fname, authenticatedUser.ID)

	err = db.Create(&d).Error

	if err != nil {
		return &Document{}, err
	}

	err = CSV2Map(file, d, db)

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

func CSV2Map(file multipart.File, d *Document, db *gorm.DB) error {
	r := csv.NewReader(file)

	var docHeaders []string

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if docHeaders == nil {
			docHeaders = record
		} else {
			dict := JSONB{}
			for i := range docHeaders {
				dict[docHeaders[i]] = record[i]
			}

			val, err := dict.Value()
			dict.Scan(val)

			details := Details{}
			if err != nil {
				return err
			}
			details.PrepareDetails(d.ID, dict)

			err = db.Create(&details).Error

			if err != nil {
				return err
			}
		}
	}

	headers, err := d.CreateHeaders(db, docHeaders)

	if err != nil {
		return err
	}

	d.Header = headers

	return nil
}

func (d *Document) CreateHeaders(db *gorm.DB, docHeaders []string) ([]Header, error) {
	headers := make([]Header, 0)

	for _, s := range docHeaders {
		h := Header{}
		h.PrepareHeader(d.ID, s)
		err := db.Create(&h).Error

		if err != nil {
			return []Header{}, err
		}

		headers = append(headers, h)
	}
	return headers, nil
}

func (d *Document) GetDocumentByID(db *gorm.DB, docID uuid.UUID) (*Document, error) {
	var err error

	err = db.Model(&Document{}).Where("id = ?", docID).Preload("Data").Take(&d).Error

	if err != nil {
		return &Document{}, err
	}

	return d, nil
}

func (d *Document) DeleteDocument(db *gorm.DB, docID uuid.UUID) (int64, error) {
	var err error

	db1 := db.Model(&Details{}).Where("document_id = ?", docID).Take(&Details{}).Delete(&Details{})
	if db1.Error != nil {
		return 0, err
	}

	db2 := db.Model(&Document{}).Where("id = ?", docID).Take(&Document{}).Delete(&Document{})

	if db2.Error != nil {
		return 0, err
	}

	return db.RowsAffected, nil
}
