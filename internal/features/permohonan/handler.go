package permohonan

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nobuww/simpel-ktp/internal/features/common"
	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
)

// Handler manages permohonan-related HTTP handlers
type Handler struct {
	store *store.Store
}

// New creates a new permohonan handler with the required dependencies
func New(s *store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

// FormData contains user data to pre-fill forms
type FormData struct {
	NIK           string
	NamaLengkap   string
	JenisKelamin  string
	Alamat        string
	NoHP          string
	Email         string
	NamaKelurahan string
	Errors        map[string]string
}

// JadwalOption represents a jadwal option for the select box
type JadwalOption struct {
	ID         string
	Label      string
	KuotaSisa  int
	StatusSesi string
}

// HandleKTPBaruForm handles the KTP Baru application form
func (h *Handler) HandleKTPBaruForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	// Fetch user profile data
	formData := h.getUserFormData(ctx, user.UserID)

	// Fetch available jadwal
	jadwalList := h.getAvailableJadwal(ctx)

	if r.Method == http.MethodGet {
		KTPBaruFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Handle POST
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		formData.Errors["general"] = "Gagal memproses form. Pastikan file tidak lebih dari 10MB."
		KTPBaruFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Get form values
	jadwalID := r.FormValue("jadwal_sesi_id")

	// Validate
	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}

	// Handle file upload
	file, header, err := r.FormFile("kartu_keluarga")
	if err != nil {
		formData.Errors["kartu_keluarga"] = "File Kartu Keluarga wajib diunggah"
	} else {
		defer file.Close()

		// Validate file type
		if !isValidFileType(header.Filename) {
			formData.Errors["kartu_keluarga"] = "Format file tidak didukung. Gunakan PDF, JPG, atau PNG"
		}
	}

	// If there are errors, re-render form
	if len(formData.Errors) > 0 {
		KTPBaruFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Upload file
	filePath, err := h.uploadFile(file, header.Filename, user.UserID, "KK")
	if err != nil {
		formData.Errors["kartu_keluarga"] = "Gagal mengunggah file"
		KTPBaruFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Create permohonan
	jadwalUUID, _ := uuid.Parse(jadwalID)
	nikText := pgtype.Text{String: user.UserID, Valid: true}
	jadwalUUIDPg := pgtype.UUID{Bytes: jadwalUUID, Valid: true}

	permohonanID, err := h.store.CreatePermohonan(ctx, pg_store.CreatePermohonanParams{
		Nik:             nikText,
		JadwalSesiID:    jadwalUUIDPg,
		JenisPermohonan: "BARU",
	})
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPBaruFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Save document reference
	permohonanUUID := pgtype.UUID{Bytes: permohonanID, Valid: true}
	err = h.store.CreateDokumenSyarat(ctx, pg_store.CreateDokumenSyaratParams{
		PermohonanID: permohonanUUID,
		FilePath:     filePath,
		JenisDokumen: "KK",
	})
	if err != nil {
		formData.Errors["general"] = "Gagal menyimpan dokumen"
		KTPBaruFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Redirect to success page
	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=baru", permohonanID.String()), http.StatusSeeOther)
}

// HandleKTPHilangForm handles the KTP Hilang application form
func (h *Handler) HandleKTPHilangForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	formData := h.getUserFormData(ctx, user.UserID)
	jadwalList := h.getAvailableJadwal(ctx)

	if r.Method == http.MethodGet {
		KTPHilangFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Handle POST
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		formData.Errors["general"] = "Gagal memproses form. Pastikan file tidak lebih dari 10MB."
		KTPHilangFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	jadwalID := r.FormValue("jadwal_sesi_id")
	nomorLaporan := r.FormValue("nomor_laporan")
	tanggalKejadian := r.FormValue("tanggal_kejadian")

	// Validate
	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}
	if nomorLaporan == "" {
		formData.Errors["nomor_laporan"] = "Nomor laporan polisi wajib diisi"
	}
	if tanggalKejadian == "" {
		formData.Errors["tanggal_kejadian"] = "Tanggal kejadian wajib diisi"
	}

	// Handle file upload
	file, header, err := r.FormFile("surat_polisi")
	if err != nil {
		formData.Errors["surat_polisi"] = "File Surat Keterangan Polisi wajib diunggah"
	} else {
		defer file.Close()
		if !isValidFileType(header.Filename) {
			formData.Errors["surat_polisi"] = "Format file tidak didukung. Gunakan PDF, JPG, atau PNG"
		}
	}

	if len(formData.Errors) > 0 {
		KTPHilangFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Upload file
	filePath, err := h.uploadFile(file, header.Filename, user.UserID, "SURAT_POLISI")
	if err != nil {
		formData.Errors["surat_polisi"] = "Gagal mengunggah file"
		KTPHilangFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Create permohonan
	jadwalUUID, _ := uuid.Parse(jadwalID)
	nikText := pgtype.Text{String: user.UserID, Valid: true}
	jadwalUUIDPg := pgtype.UUID{Bytes: jadwalUUID, Valid: true}

	permohonanID, err := h.store.CreatePermohonan(ctx, pg_store.CreatePermohonanParams{
		Nik:             nikText,
		JadwalSesiID:    jadwalUUIDPg,
		JenisPermohonan: "HILANG",
	})
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPHilangFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Save document reference
	permohonanUUID := pgtype.UUID{Bytes: permohonanID, Valid: true}
	err = h.store.CreateDokumenSyarat(ctx, pg_store.CreateDokumenSyaratParams{
		PermohonanID: permohonanUUID,
		FilePath:     filePath,
		JenisDokumen: "SURAT_POLISI",
	})
	if err != nil {
		formData.Errors["general"] = "Gagal menyimpan dokumen"
		KTPHilangFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=hilang", permohonanID.String()), http.StatusSeeOther)
}

// HandleKTPRusakForm handles the KTP Rusak application form
func (h *Handler) HandleKTPRusakForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	formData := h.getUserFormData(ctx, user.UserID)
	jadwalList := h.getAvailableJadwal(ctx)

	if r.Method == http.MethodGet {
		KTPRusakFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Handle POST
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB for multiple files
		formData.Errors["general"] = "Gagal memproses form. Pastikan file tidak lebih dari 20MB."
		KTPRusakFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	jadwalID := r.FormValue("jadwal_sesi_id")
	deskripsiKerusakan := r.FormValue("deskripsi_kerusakan")

	// Validate
	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}
	if deskripsiKerusakan == "" {
		formData.Errors["deskripsi_kerusakan"] = "Deskripsi kerusakan wajib diisi"
	}

	// Handle KTP file
	ktpFile, ktpHeader, err := r.FormFile("ktp_rusak")
	if err != nil {
		formData.Errors["ktp_rusak"] = "Foto KTP rusak wajib diunggah"
	} else {
		defer ktpFile.Close()
		if !isValidFileType(ktpHeader.Filename) {
			formData.Errors["ktp_rusak"] = "Format file tidak didukung. Gunakan PDF, JPG, atau PNG"
		}
	}

	// Handle KK file
	kkFile, kkHeader, err := r.FormFile("kartu_keluarga")
	if err != nil {
		formData.Errors["kartu_keluarga"] = "File Kartu Keluarga wajib diunggah"
	} else {
		defer kkFile.Close()
		if !isValidFileType(kkHeader.Filename) {
			formData.Errors["kartu_keluarga"] = "Format file tidak didukung. Gunakan PDF, JPG, atau PNG"
		}
	}

	if len(formData.Errors) > 0 {
		KTPRusakFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Upload files
	ktpPath, err := h.uploadFile(ktpFile, ktpHeader.Filename, user.UserID, "KTP_RUSAK")
	if err != nil {
		formData.Errors["ktp_rusak"] = "Gagal mengunggah file KTP"
		KTPRusakFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	kkPath, err := h.uploadFile(kkFile, kkHeader.Filename, user.UserID, "KK")
	if err != nil {
		formData.Errors["kartu_keluarga"] = "Gagal mengunggah file KK"
		KTPRusakFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Create permohonan
	jadwalUUID, _ := uuid.Parse(jadwalID)
	nikText := pgtype.Text{String: user.UserID, Valid: true}
	jadwalUUIDPg := pgtype.UUID{Bytes: jadwalUUID, Valid: true}

	permohonanID, err := h.store.CreatePermohonan(ctx, pg_store.CreatePermohonanParams{
		Nik:             nikText,
		JadwalSesiID:    jadwalUUIDPg,
		JenisPermohonan: "RUSAK",
	})
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPRusakFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Save documents
	permohonanUUID := pgtype.UUID{Bytes: permohonanID, Valid: true}
	h.store.CreateDokumenSyarat(ctx, pg_store.CreateDokumenSyaratParams{
		PermohonanID: permohonanUUID,
		FilePath:     ktpPath,
		JenisDokumen: "KTP_RUSAK",
	})
	h.store.CreateDokumenSyarat(ctx, pg_store.CreateDokumenSyaratParams{
		PermohonanID: permohonanUUID,
		FilePath:     kkPath,
		JenisDokumen: "KK",
	})

	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=rusak", permohonanID.String()), http.StatusSeeOther)
}

// HandleKTPUbahForm handles the Perubahan Data KTP application form
func (h *Handler) HandleKTPUbahForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	formData := h.getUserFormData(ctx, user.UserID)
	jadwalList := h.getAvailableJadwal(ctx)

	if r.Method == http.MethodGet {
		KTPUbahFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Handle POST
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		formData.Errors["general"] = "Gagal memproses form. Pastikan file tidak lebih dari 20MB."
		KTPUbahFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	jadwalID := r.FormValue("jadwal_sesi_id")
	alasanPerubahan := r.FormValue("alasan_perubahan")

	// Validate
	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}
	if alasanPerubahan == "" {
		formData.Errors["alasan_perubahan"] = "Alasan perubahan wajib diisi"
	}

	// Handle KTP file
	ktpFile, ktpHeader, err := r.FormFile("ktp_lama")
	if err != nil {
		formData.Errors["ktp_lama"] = "Foto KTP lama wajib diunggah"
	} else {
		defer ktpFile.Close()
		if !isValidFileType(ktpHeader.Filename) {
			formData.Errors["ktp_lama"] = "Format file tidak didukung. Gunakan PDF, JPG, atau PNG"
		}
	}

	// Handle KK file
	kkFile, kkHeader, err := r.FormFile("kartu_keluarga")
	if err != nil {
		formData.Errors["kartu_keluarga"] = "File Kartu Keluarga wajib diunggah"
	} else {
		defer kkFile.Close()
		if !isValidFileType(kkHeader.Filename) {
			formData.Errors["kartu_keluarga"] = "Format file tidak didukung. Gunakan PDF, JPG, atau PNG"
		}
	}

	if len(formData.Errors) > 0 {
		KTPUbahFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Upload files
	ktpPath, err := h.uploadFile(ktpFile, ktpHeader.Filename, user.UserID, "KTP")
	if err != nil {
		formData.Errors["ktp_lama"] = "Gagal mengunggah file KTP"
		KTPUbahFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	kkPath, err := h.uploadFile(kkFile, kkHeader.Filename, user.UserID, "KK")
	if err != nil {
		formData.Errors["kartu_keluarga"] = "Gagal mengunggah file KK"
		KTPUbahFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Create permohonan
	jadwalUUID, _ := uuid.Parse(jadwalID)
	nikText := pgtype.Text{String: user.UserID, Valid: true}
	jadwalUUIDPg := pgtype.UUID{Bytes: jadwalUUID, Valid: true}

	permohonanID, err := h.store.CreatePermohonan(ctx, pg_store.CreatePermohonanParams{
		Nik:             nikText,
		JadwalSesiID:    jadwalUUIDPg,
		JenisPermohonan: "UPDATE",
	})
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPUbahFormPage(formData, jadwalList).Render(ctx, w)
		return
	}

	// Save documents
	permohonanUUID := pgtype.UUID{Bytes: permohonanID, Valid: true}
	h.store.CreateDokumenSyarat(ctx, pg_store.CreateDokumenSyaratParams{
		PermohonanID: permohonanUUID,
		FilePath:     ktpPath,
		JenisDokumen: "KTP",
	})
	h.store.CreateDokumenSyarat(ctx, pg_store.CreateDokumenSyaratParams{
		PermohonanID: permohonanUUID,
		FilePath:     kkPath,
		JenisDokumen: "KK",
	})

	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=ubah", permohonanID.String()), http.StatusSeeOther)
}

// HandleSuccessPage shows the success page after form submission
func (h *Handler) HandleSuccessPage(w http.ResponseWriter, r *http.Request) {
	_, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	permohonanID := r.URL.Query().Get("id")
	applicationType := r.URL.Query().Get("type")

	if permohonanID == "" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Get permohonan details
	permohonanUUID, _ := uuid.Parse(permohonanID)
	detail, err := h.store.GetPermohonanDetail(ctx, permohonanUUID)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	successData := SuccessData{
		PermohonanID:    permohonanID,
		KodeBooking:     detail.KodeBooking.String,
		ApplicationType: formatApplicationType(applicationType),
		JadwalTanggal:   formatDate(detail.JadwalTanggal),
		JadwalJam:       formatTime(detail.JadwalJamMulai) + " - " + formatTime(detail.JadwalJamSelesai),
		NamaKelurahan:   detail.NamaKelurahan.String,
	}

	SuccessPage(successData).Render(ctx, w)
}

// Helper functions

func (h *Handler) getUserFormData(ctx context.Context, nik string) FormData {
	formData := FormData{
		NIK:    nik,
		Errors: make(map[string]string),
	}

	profile, err := h.store.GetPendudukProfile(ctx, nik)
	if err == nil {
		formData.NamaLengkap = profile.NamaLengkap
		formData.JenisKelamin = profile.JenisKelamin
		formData.Alamat = profile.Alamat.String
		formData.NoHP = profile.NoHp.String
		formData.Email = profile.Email.String
		formData.NamaKelurahan = profile.NamaKelurahan.String
	}

	return formData
}

func (h *Handler) getAvailableJadwal(ctx context.Context) []JadwalOption {
	// Get jadwal from today onwards
	today := time.Now()
	nextMonth := today.AddDate(0, 1, 0)

	todayPg := pgtype.Date{Time: today, Valid: true}
	nextMonthPg := pgtype.Date{Time: nextMonth, Valid: true}

	jadwalList, err := h.store.ListJadwalSesi(ctx, pg_store.ListJadwalSesiParams{
		Tanggal:   todayPg,
		Tanggal_2: nextMonthPg,
	})
	if err != nil {
		return nil
	}

	options := make([]JadwalOption, 0, len(jadwalList))
	for _, j := range jadwalList {
		if j.StatusSesi.String != "BUKA" {
			continue
		}
		kuotaSisa := int(j.KuotaMaksimal - j.KuotaTerisi)
		if kuotaSisa <= 0 {
			continue
		}

		// Format: "Senin, 01 Jan 2024 - 08:00 (Kelurahan X, sisa 10 kuota)"
		label := fmt.Sprintf("%s - %s (%s, sisa %d kuota)",
			formatDate(j.Tanggal),
			formatTime(j.JamMulai),
			j.NamaKelurahan,
			kuotaSisa,
		)

		options = append(options, JadwalOption{
			ID:         j.ID.String(),
			Label:      label,
			KuotaSisa:  kuotaSisa,
			StatusSesi: j.StatusSesi.String,
		})
	}

	return options
}

func (h *Handler) uploadFile(file io.Reader, filename, userID, docType string) (string, error) {
	// Create upload directory
	uploadDir := filepath.Join("static", "uploads", userID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%s_%s_%d%s", docType, userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, newFilename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filePath, nil
}

func isValidFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".pdf", ".jpg", ".jpeg", ".png"}
	for _, v := range validExts {
		if ext == v {
			return true
		}
	}
	return false
}

func formatApplicationType(typeCode string) string {
	switch typeCode {
	case "baru":
		return "KTP Baru"
	case "hilang":
		return "KTP Hilang"
	case "rusak":
		return "KTP Rusak"
	case "ubah":
		return "Perubahan Data KTP"
	default:
		return typeCode
	}
}

func formatDate(d pgtype.Date) string {
	if !d.Valid {
		return "-"
	}
	return d.Time.Format("02 Jan 2006")
}

func formatTime(t pgtype.Time) string {
	if !t.Valid {
		return "-"
	}
	hour := t.Microseconds / 3600000000
	minute := (t.Microseconds % 3600000000) / 60000000
	return fmt.Sprintf("%02d:%02d", hour, minute)
}
