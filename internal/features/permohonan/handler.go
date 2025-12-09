package permohonan

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/nobuww/simpel-ktp/internal/features/common"
)

type Handler struct {
	service Service
}

func New(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) HandleKTPBaruForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	formData, err := h.service.GetFormData(ctx, user.UserID)
	if err != nil {
		http.Error(w, "Gagal memuat data user", http.StatusInternalServerError)
		return
	}

	locations, _ := h.service.GetLocations(ctx)
	var jadwalList []JadwalOption
	if lokasiIDStr := r.FormValue("lokasi_id"); lokasiIDStr != "" {
		if lid, err := strconv.Atoi(lokasiIDStr); err == nil {
			val := int32(lid)
			jadwalList, _ = h.service.GetAvailableJadwal(ctx, &val)
		}
	}

	if r.Method == http.MethodGet {
		KTPBaruFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	// Parse multipart form data (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		formData.Errors["general"] = "Gagal memproses form: " + err.Error()
		KTPBaruFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	jadwalID := r.FormValue("jadwal_sesi_id")

	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}

	filePath, err := common.ProcessUpload(r, "kartu_keluarga", user.UserID, "KK")
	if err != nil {
		formData.Errors["kartu_keluarga"] = err.Error()
	}

	if len(formData.Errors) > 0 {
		KTPBaruFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	req := CreatePermohonanRequest{
		UserID:   user.UserID,
		JadwalID: jadwalID,
		Type:     "baru",
		Documents: []DocumentFile{
			{Type: "KK", Path: filePath},
		},
	}

	permohonanID, err := h.service.CreatePermohonan(ctx, req)
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPBaruFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=baru", permohonanID.String()), http.StatusSeeOther)
}

func (h *Handler) HandleKTPHilangForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	formData, err := h.service.GetFormData(ctx, user.UserID)
	if err != nil {
		http.Error(w, "Gagal memuat data user", http.StatusInternalServerError)
		return
	}
	locations, _ := h.service.GetLocations(ctx)
	var jadwalList []JadwalOption
	if lokasiIDStr := r.FormValue("lokasi_id"); lokasiIDStr != "" {
		if lid, err := strconv.Atoi(lokasiIDStr); err == nil {
			val := int32(lid)
			jadwalList, _ = h.service.GetAvailableJadwal(ctx, &val)
		}
	}

	if r.Method == http.MethodGet {
		KTPHilangFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	// Parse multipart form data (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		formData.Errors["general"] = "Gagal memproses form: " + err.Error()
		KTPHilangFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	jadwalID := r.FormValue("jadwal_sesi_id")
	nomorLaporan := r.FormValue("nomor_laporan")
	tanggalKejadian := r.FormValue("tanggal_kejadian")

	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}
	if nomorLaporan == "" {
		formData.Errors["nomor_laporan"] = "Nomor laporan polisi wajib diisi"
	}
	if tanggalKejadian == "" {
		formData.Errors["tanggal_kejadian"] = "Tanggal kejadian wajib diisi"
	}

	filePath, err := common.ProcessUpload(r, "surat_polisi", user.UserID, "SURAT_POLISI")
	if err != nil {
		formData.Errors["surat_polisi"] = err.Error()
	}

	if len(formData.Errors) > 0 {
		KTPHilangFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	req := CreatePermohonanRequest{
		UserID:   user.UserID,
		JadwalID: jadwalID,
		Type:     "hilang",
		Documents: []DocumentFile{
			{Type: "SURAT_POLISI", Path: filePath},
		},
	}

	permohonanID, err := h.service.CreatePermohonan(ctx, req)
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPHilangFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=hilang", permohonanID.String()), http.StatusSeeOther)
}

func (h *Handler) HandleKTPRusakForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	formData, err := h.service.GetFormData(ctx, user.UserID)
	if err != nil {
		http.Error(w, "Gagal memuat data user", http.StatusInternalServerError)
		return
	}
	locations, _ := h.service.GetLocations(ctx)
	var jadwalList []JadwalOption
	if lokasiIDStr := r.FormValue("lokasi_id"); lokasiIDStr != "" {
		if lid, err := strconv.Atoi(lokasiIDStr); err == nil {
			val := int32(lid)
			jadwalList, _ = h.service.GetAvailableJadwal(ctx, &val)
		}
	}

	if r.Method == http.MethodGet {
		KTPRusakFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	// Parse multipart form data (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		formData.Errors["general"] = "Gagal memproses form: " + err.Error()
		KTPRusakFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	jadwalID := r.FormValue("jadwal_sesi_id")
	deskripsiKerusakan := r.FormValue("deskripsi_kerusakan")

	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}
	if deskripsiKerusakan == "" {
		formData.Errors["deskripsi_kerusakan"] = "Deskripsi kerusakan wajib diisi"
	}

	ktpPath, err := common.ProcessUpload(r, "ktp_rusak", user.UserID, "KTP_RUSAK")
	if err != nil {
		formData.Errors["ktp_rusak"] = err.Error()
	}

	kkPath, err := common.ProcessUpload(r, "kartu_keluarga", user.UserID, "KK")
	if err != nil {
		formData.Errors["kartu_keluarga"] = err.Error()
	}

	if len(formData.Errors) > 0 {
		KTPRusakFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	req := CreatePermohonanRequest{
		UserID:   user.UserID,
		JadwalID: jadwalID,
		Type:     "rusak",
		Documents: []DocumentFile{
			{Type: "KTP_RUSAK", Path: ktpPath},
			{Type: "KK", Path: kkPath},
		},
	}

	permohonanID, err := h.service.CreatePermohonan(ctx, req)
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPRusakFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=rusak", permohonanID.String()), http.StatusSeeOther)
}

func (h *Handler) HandleKTPUbahForm(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	formData, err := h.service.GetFormData(ctx, user.UserID)
	if err != nil {
		http.Error(w, "Gagal memuat data user", http.StatusInternalServerError)
		return
	}
	locations, _ := h.service.GetLocations(ctx)
	var jadwalList []JadwalOption
	if lokasiIDStr := r.FormValue("lokasi_id"); lokasiIDStr != "" {
		if lid, err := strconv.Atoi(lokasiIDStr); err == nil {
			val := int32(lid)
			jadwalList, _ = h.service.GetAvailableJadwal(ctx, &val)
		}
	}

	if r.Method == http.MethodGet {
		KTPUbahFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	// Parse multipart form data (10MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		formData.Errors["general"] = "Gagal memproses form: " + err.Error()
		KTPUbahFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	jadwalID := r.FormValue("jadwal_sesi_id")
	alasanPerubahan := r.FormValue("alasan_perubahan")

	if jadwalID == "" {
		formData.Errors["jadwal_sesi_id"] = "Pilih jadwal kedatangan"
	}
	if alasanPerubahan == "" {
		formData.Errors["alasan_perubahan"] = "Alasan perubahan wajib diisi"
	}

	ktpPath, err := common.ProcessUpload(r, "ktp_lama", user.UserID, "KTP")
	if err != nil {
		formData.Errors["ktp_lama"] = err.Error()
	}

	kkPath, err := common.ProcessUpload(r, "kartu_keluarga", user.UserID, "KK")
	if err != nil {
		formData.Errors["kartu_keluarga"] = err.Error()
	}

	if len(formData.Errors) > 0 {
		KTPUbahFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	req := CreatePermohonanRequest{
		UserID:   user.UserID,
		JadwalID: jadwalID,
		Type:     "ubah",
		Documents: []DocumentFile{
			{Type: "KTP", Path: ktpPath},
			{Type: "KK", Path: kkPath},
		},
	}

	permohonanID, err := h.service.CreatePermohonan(ctx, req)
	if err != nil {
		formData.Errors["general"] = "Gagal membuat permohonan: " + err.Error()
		KTPUbahFormPage(formData, locations, jadwalList).Render(ctx, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/permohonan/sukses?id=%s&type=ubah", permohonanID.String()), http.StatusSeeOther)
}

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

	successData, err := h.service.GetSuccessData(ctx, permohonanID, applicationType)
	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	SuccessPage(successData).Render(ctx, w)
}

func (h *Handler) HandleGetJadwalOptions(w http.ResponseWriter, r *http.Request) {
	lokasiIDStr := r.URL.Query().Get("lokasi_id")
	if lokasiIDStr == "" {
		JadwalSelectPartial([]JadwalOption{}, "").Render(r.Context(), w)
		return
	}

	lokasiID, err := strconv.Atoi(lokasiIDStr)
	if err != nil {
		JadwalSelectPartial([]JadwalOption{}, "Invalid Location ID").Render(r.Context(), w)
		return
	}

	lid := int32(lokasiID)
	jadwalList, err := h.service.GetAvailableJadwal(r.Context(), &lid)
	if err != nil {
		JadwalSelectPartial([]JadwalOption{}, "Error fetching schedules").Render(r.Context(), w)
		return
	}

	JadwalSelectPartial(jadwalList, "").Render(r.Context(), w)
}
