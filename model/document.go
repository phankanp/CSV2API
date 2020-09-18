package model

import (
	"encoding/csv"
	"encoding/json"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/pborman/uuid"
)

// Maps csv file row data
type JSONB map[string]interface{}

// CSV file model
type Document struct {
	ID        uuid.UUID `gorm:"primary_key;" json:"id"`
	UserID    uuid.UUID `json:"-"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Header    []Header  `gorm:"not null" json:"headers"`
	Row       []Row     `gorm:"OnDelete:SET NULL;" json:"rows"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
}

// CSV row model
type Row struct {
	ID         uint           `gorm:"primary_key;auto_increment" json:"id"`
	DocumentID uuid.UUID      `gorm:"not null" json:"-"`
	Data       datatypes.JSON `type:"jsonb not null default '{}'::jsonb" json:"data"`
	CreatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"-"`
}

// CSV header model
type Header struct {
	ID         uint      `gorm:"primary_key;auto_increment" json:"-"`
	DocumentID uuid.UUID `gorm:"not null" json:"-"`
	Name       string    `gorm:"not null" json:"name"`
}

// Assign data to document model
func (d *Document) PrepareDocument(fname string, uid uuid.UUID) {
	d.ID = uuid.NewRandom()
	d.UserID = uid
	d.Title = strings.TrimSpace(fname)
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
}

// Assign data to row model
func (r *Row) PrepareRow(docID uuid.UUID, data datatypes.JSON) {
	r.DocumentID = docID
	r.Data = data
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
}

// Assign data to header model
func (h *Header) PrepareHeader(docID uuid.UUID, name string) {
	h.DocumentID = docID
	h.Name = name
}

// Creates a document in database
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

	rows := []Row{}

	err = db.Model(&Row{}).Where("document_id = ?", d.ID).Find(&rows).Error

	if err != nil {
		return &Document{}, err
	}

	d.Row = rows

	return d, nil
}

// Converts csv rows to JSONB map and creates document headers
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

			j, err := json.Marshal(dict)

			if err != nil {
				return err
			}

			rows := Row{}

			rows.PrepareRow(d.ID, j)

			err = db.Create(&rows).Error

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

// Creates document headers in database
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

// Gets all document for a user
func (d *Document) GetDocuments(db *gorm.DB, uid uuid.UUID) (*[]Document, error) {
	documents := []Document{}

	err := db.Model(&Document{}).Where("user_id = ?", uid).Preload("Row").Preload("Header").Find(&documents).Error

	if err != nil {
		return &[]Document{}, err
	}

	return &documents, nil
}

// Gets a document by id
func (d *Document) GetDocumentByID(db *gorm.DB, docID uuid.UUID) (*Document, error) {
	var err error

	err = db.Model(&Document{}).Where("id = ?", docID).Preload("Row").Take(&d).Error

	if err != nil {
		return &Document{}, err
	}

	return d, nil
}

// Deletes a document
func (d *Document) DeleteDocument(db *gorm.DB, docID uuid.UUID) (int64, error) {
	var err error

	dbRow := db.Model(&Row{}).Where("document_id = ?", docID).Take(&Row{}).Delete(&Row{})

	if dbRow.Error != nil {
		return 0, err
	}

	dbHeader := db.Model(&Header{}).Where("document_id = ?", docID).Take(&Header{}).Delete(&Header{})

	if dbHeader.Error != nil {
		return 0, err
	}

	dbDocument := db.Model(&Document{}).Where("id = ?", docID).Take(&Document{}).Delete(&Document{})

	if dbDocument.Error != nil {
		return 0, err
	}

	return db.RowsAffected, nil
}

// Gets headers for a document
func (d *Document) GetDocumentHeaders(db *gorm.DB) ([]Header, error) {
	headers := []Header{}

	err := db.Model(&Header{}).Where("document_id = ?", d.ID).Find(&headers).Error

	if err != nil {
		return []Header{}, err
	}

	return headers, nil
}

// Gets all rows for a document
func (r *Row) GetAllRowsByDocument(db *gorm.DB, docID uuid.UUID) (*[]Row, error) {
	rows := []Row{}

	err := db.Model(&Row{}).Where("document_id = ?", docID).Find(&rows).Error

	if err != nil {
		return &[]Row{}, err
	}

	return &rows, nil
}

// Gets specified row for a document
func (r *Row) GetRowByID(db *gorm.DB, docID uuid.UUID, rowID uint) (*Row, error) {
	err := db.Model(&Row{}).Where("document_id = ? AND id = ?", docID, rowID).Take(&r).Error

	if err != nil {
		return &Row{}, err
	}

	return r, nil
}

// Creates a new row in a document
func (r *Row) CreateRow(db *gorm.DB, docID uuid.UUID, rowData JSONB) (*Row, error) {
	j, err := json.Marshal(rowData)

	if err != nil {
		return &Row{}, err
	}

	r.PrepareRow(docID, j)

	err = db.Create(&r).Error

	if err != nil {
		return &Row{}, err
	}

	return r, nil
}

// Updates a row in a document
func (r *Row) UpdateRow(db *gorm.DB, rowData JSONB) (*Row, error) {
	j, err := json.Marshal(rowData)

	if err != nil {
		return &Row{}, err
	}

	err = db.Model(&Row{}).Where("id = ?", r.ID).Updates(Row{Data: j, UpdatedAt: time.Now()}).Error

	if err != nil {
		return &Row{}, err
	}

	r.Data = j

	return r, nil
}

// Deletes a row in a document
func (r *Row) DeleteRow(db *gorm.DB, docID uuid.UUID, rowID uint) (int64, error) {
	dbRow := db.Model(&Row{}).Where("document_id = ? AND id = ?", docID, rowID).Take(&Row{}).Delete(&Row{})

	if dbRow.Error != nil {
		return 0, db.Error
	}

	return db.RowsAffected, nil
}

// Searches rows in a documents and return rows matching specified parameters
func (r *Row) SearchRows(db *gorm.DB, docID uuid.UUID, headerInput string, dataInput string) (*[]Row, error) {
	rows := []Row{}

	err := db.Model(&Row{}).Where("document_id = ?", docID).Find(&rows, datatypes.JSONQuery("data").Equals(dataInput, headerInput)).Error

	if err != nil {
		return &[]Row{}, err
	}

	return &rows, nil
}
