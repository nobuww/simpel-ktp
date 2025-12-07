package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxFileSize = 10 << 20 // 10MB
)

func ProcessUpload(r *http.Request, formKey string, userID string, docType string) (string, error) {
	file, header, err := r.FormFile(formKey)
	if err != nil {
		return "", fmt.Errorf("file %s wajib diunggah", formKey)
	}
	defer file.Close()

	if header.Size > MaxFileSize {
		return "", fmt.Errorf("ukuran file melebihi batas 10MB")
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("gagal membaca file")
	}
	
	contentType := http.DetectContentType(buffer[:n])

	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowedTypes[contentType] {
		return "", fmt.Errorf("format file tidak didukung. Gunakan PDF, JPG, atau PNG (terdeteksi: %s)", contentType)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("gagal memproses file")
	}

	uploadDir := filepath.Join("static", "uploads", userID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori upload")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	
	validExt := false
	switch contentType {
	case "image/jpeg":
		if ext == ".jpg" || ext == ".jpeg" { validExt = true }
	case "image/png":
		if ext == ".png" { validExt = true }
	case "application/pdf":
		if ext == ".pdf" { validExt = true }
	}

	if !validExt {
		return "", fmt.Errorf("ekstensi file tidak sesuai dengan konten (tipe: %s)", contentType)
	}

	newFilename := fmt.Sprintf("%s_%s_%d%s", docType, userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, newFilename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("gagal menyalin file")
	}

	return filePath, nil
}
