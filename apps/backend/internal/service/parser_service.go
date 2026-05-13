package service

import (
	"encoding/csv"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	apperrors "github.com/devrapture/omni/internal/errors"
	"github.com/fumiama/go-docx"
)

type ParserService struct{}

func NewParserService() *ParserService {
	return &ParserService{}
}

func (s *ParserService) Parse(path string) (text, sourceType string, err error) {
	ext := strings.ToLower(filepath.Ext(path))

	src, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	head := make([]byte, 512)
	n, _ := src.Read(head)
	mime := http.DetectContentType(head[:n])

	if !isSupportedParserContent(ext, mime) {
		return "", "", apperrors.ErrNotSupportFile
	}

	switch ext {
	case ".csv":
		text, err = s.parseCSV(path)
		if err != nil {
			return "", "", err
		}
		sourceType = ".csv"
	case ".docx":
		text, err = s.parseDocx(path)
		if err != nil {
			return "", "", err
		}
		sourceType = ".docx"
	default:
		return "", "", apperrors.ErrNotSupportFile
	}

	return text, sourceType, nil
}

func isSupportedParserContent(ext, mime string) bool {
	switch ext {
	case ".csv":
		return mime == "text/csv" || strings.HasPrefix(mime, "text/plain")
	case ".pdf":
		return mime == "application/pdf"
	case ".xlsx", ".docx":
		return mime == "application/zip" ||
			mime == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" ||
			mime == "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return false
	}
}

func (s *ParserService) parseCSV(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return "", err
	}

	if len(records) == 0 {
		return "", apperrors.ErrEmptyCsvFile
	}

	var text strings.Builder
	for _, record := range records {
		if len(record) == 0 {
			continue
		}
		text.WriteString(strings.Join(record, " | "))
		text.WriteByte('\n')
	}

	parsedText := strings.TrimSpace(text.String())
	if parsedText == "" {
		return "", apperrors.ErrEmptyCsvFile
	}

	return parsedText, nil
}

func (s *ParserService) parseDocx(path string) (string, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer readFile.Close()

	fileinfo, err := readFile.Stat()
	if err != nil {
		return "", err
	}

	size := fileinfo.Size()
	doc, err := docx.Parse(readFile, size)
	if err != nil {
		return "", err
	}

	var text strings.Builder
	for _, it := range doc.Document.Body.Items {
		switch item := it.(type) {
		case *docx.Paragraph:
			writeDocxText(&text, item.String())
		case *docx.Table:
			writeDocxText(&text, item.String())
		}
	}

	parsedText := strings.TrimSpace(text.String())
	if parsedText == "" {
		return "", apperrors.ErrEmptyDocxFile
	}

	return parsedText, nil
}

func writeDocxText(text *strings.Builder, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	if text.Len() > 0 {
		text.WriteByte('\n')
	}
	text.WriteString(value)
}
