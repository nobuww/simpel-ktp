package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nobuww/simpel-ktp/internal/middleware"
	"github.com/nobuww/simpel-ktp/internal/session"
	"github.com/nobuww/simpel-ktp/internal/store"
)

// Handler manages admin-related HTTP handlers
type Handler struct {
	store *store.Store
}

// New creates a new admin handler with the required dependencies
func New(s *store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/petugas/login", http.StatusSeeOther)
		return
	}

	data := DashboardData{
		UserName:   user.UserName,
		UserRole:   formatRole(user.UserRole),
		ActivePage: "dashboard",
	}

	DashboardPage(data).Render(r.Context(), w)
}

func (h *Handler) PermohonanHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/petugas/login", http.StatusSeeOther)
		return
	}

	// Mock data for now - will be replaced with actual DB queries
	data := PermohonanPageData{
		UserName:   user.UserName,
		UserRole:   formatRole(user.UserRole),
		ActivePage: "permohonan",
		Stats: PermohonanStats{
			Total:      63,
			Verifikasi: 12,
			Proses:     28,
			SiapAmbil:  8,
			Selesai:    12,
			Ditolak:    3,
		},
		List: []PermohonanItem{
			{ID: "1", KodeBooking: "JKT-02-A1B2", NIK: "3201234567890001", NamaLengkap: "Ahmad Wijaya", JenisPermohonan: "BARU", StatusTerkini: "VERIFIKASI", TanggalDaftar: "2025-12-01", JadwalSesi: "Senin, 2 Des 08:00", NomorAntrian: 1},
			{ID: "2", KodeBooking: "JKT-02-C3D4", NIK: "3201234567890002", NamaLengkap: "Siti Nurhaliza", JenisPermohonan: "HILANG", StatusTerkini: "PROSES", TanggalDaftar: "2025-12-01", JadwalSesi: "Senin, 2 Des 08:00", NomorAntrian: 2},
			{ID: "3", KodeBooking: "JKT-02-E5F6", NIK: "3201234567890003", NamaLengkap: "Budi Santoso", JenisPermohonan: "UPDATE", StatusTerkini: "SIAP_AMBIL", TanggalDaftar: "2025-11-30", JadwalSesi: "Senin, 2 Des 10:00", NomorAntrian: 3},
			{ID: "4", KodeBooking: "JKT-02-G7H8", NIK: "3201234567890004", NamaLengkap: "Dewi Lestari", JenisPermohonan: "RUSAK", StatusTerkini: "SELESAI", TanggalDaftar: "2025-11-29", JadwalSesi: "Jumat, 29 Nov 13:00", NomorAntrian: 15},
			{ID: "5", KodeBooking: "JKT-02-I9J0", NIK: "3201234567890005", NamaLengkap: "Rudi Hermawan", JenisPermohonan: "BARU", StatusTerkini: "DITOLAK", TanggalDaftar: "2025-11-28", JadwalSesi: "Kamis, 28 Nov 08:00", NomorAntrian: 5},
		},
	}

	PermohonanPage(data).Render(r.Context(), w)
}

// Mock data store - will be replaced with actual DB queries
var mockPermohonanStatus = map[string]string{
	"1": "VERIFIKASI",
	"2": "PROSES",
	"3": "SIAP_AMBIL",
	"4": "SELESAI",
	"5": "DITOLAK",
}

func (h *Handler) PermohonanDetailHandler(w http.ResponseWriter, r *http.Request) {
	permohonanID := chi.URLParam(r, "id")

	// Mock data - will be replaced with actual DB query
	detailMap := map[string]PermohonanDetail{
		"1": {
			ID:               "1",
			KodeBooking:      "JKT-02-A1B2",
			NIK:              "3201234567890001",
			NamaLengkap:      "Ahmad Wijaya",
			TempatLahir:      "Jakarta",
			TanggalLahir:     "15 Maret 1990",
			JenisKelamin:     "Laki-laki",
			Alamat:           "Jl. Menteng Raya No. 45",
			RT:               "005",
			RW:               "012",
			Kelurahan:        "Menteng",
			Kecamatan:        "Menteng",
			Agama:            "Islam",
			StatusPerkawinan: "Kawin",
			Pekerjaan:        "Karyawan Swasta",
			Kewarganegaraan:  "WNI",
			NoTelp:           "081234567890",
			JenisPermohonan:  "BARU",
			AlasanPermohonan: "Pembuatan KTP untuk pertama kali setelah berusia 17 tahun.",
			StatusTerkini:    "VERIFIKASI",
			TanggalDaftar:    "1 Desember 2025",
			JadwalSesi:       "Senin, 2 Des 2025 - 08:00 WIB",
			NomorAntrian:     1,
			Catatan:          "",
			RiwayatStatus: []RiwayatStatusItem{
				{Status: "VERIFIKASI", Waktu: "1 Des 2025, 14:30", Petugas: "Admin Kelurahan", Catatan: "Menunggu verifikasi dokumen"},
				{Status: "DAFTAR", Waktu: "1 Des 2025, 10:15", Petugas: "Sistem", Catatan: "Permohonan berhasil didaftarkan"},
			},
		},
		"2": {
			ID:               "2",
			KodeBooking:      "JKT-02-C3D4",
			NIK:              "3201234567890002",
			NamaLengkap:      "Siti Nurhaliza",
			TempatLahir:      "Bandung",
			TanggalLahir:     "22 Juli 1985",
			JenisKelamin:     "Perempuan",
			Alamat:           "Jl. Cikini Raya No. 88",
			RT:               "003",
			RW:               "008",
			Kelurahan:        "Cikini",
			Kecamatan:        "Menteng",
			Agama:            "Islam",
			StatusPerkawinan: "Kawin",
			Pekerjaan:        "Ibu Rumah Tangga",
			Kewarganegaraan:  "WNI",
			NoTelp:           "082345678901",
			JenisPermohonan:  "HILANG",
			AlasanPermohonan: "KTP hilang saat bepergian, sudah membuat laporan kehilangan di kepolisian.",
			StatusTerkini:    "PROSES",
			TanggalDaftar:    "1 Desember 2025",
			JadwalSesi:       "Senin, 2 Des 2025 - 08:00 WIB",
			NomorAntrian:     2,
			Catatan:          "Dokumen surat kehilangan sudah diverifikasi",
			RiwayatStatus: []RiwayatStatusItem{
				{Status: "PROSES", Waktu: "1 Des 2025, 16:00", Petugas: "Budi Santoso", Catatan: "Dokumen lengkap, proses pencetakan"},
				{Status: "VERIFIKASI", Waktu: "1 Des 2025, 14:30", Petugas: "Admin Kelurahan", Catatan: "Memverifikasi surat kehilangan"},
				{Status: "DAFTAR", Waktu: "1 Des 2025, 09:00", Petugas: "Sistem", Catatan: "Permohonan berhasil didaftarkan"},
			},
		},
		"3": {
			ID:               "3",
			KodeBooking:      "JKT-02-E5F6",
			NIK:              "3201234567890003",
			NamaLengkap:      "Budi Santoso",
			TempatLahir:      "Surabaya",
			TanggalLahir:     "5 Januari 1978",
			JenisKelamin:     "Laki-laki",
			Alamat:           "Jl. Gondangdia Lama No. 12",
			RT:               "001",
			RW:               "004",
			Kelurahan:        "Gondangdia",
			Kecamatan:        "Menteng",
			Agama:            "Kristen",
			StatusPerkawinan: "Kawin",
			Pekerjaan:        "Wiraswasta",
			Kewarganegaraan:  "WNI",
			NoTelp:           "083456789012",
			JenisPermohonan:  "UPDATE",
			AlasanPermohonan: "Perubahan alamat tempat tinggal karena pindah rumah.",
			StatusTerkini:    "SIAP_AMBIL",
			TanggalDaftar:    "30 November 2025",
			JadwalSesi:       "Senin, 2 Des 2025 - 10:00 WIB",
			NomorAntrian:     3,
			Catatan:          "KTP sudah siap, silakan ambil di loket dengan membawa bukti pendaftaran",
			RiwayatStatus: []RiwayatStatusItem{
				{Status: "SIAP_AMBIL", Waktu: "1 Des 2025, 10:00", Petugas: "Admin Kecamatan", Catatan: "KTP sudah dicetak dan siap diambil"},
				{Status: "PROSES", Waktu: "30 Nov 2025, 15:00", Petugas: "Budi Santoso", Catatan: "Proses pencetakan KTP"},
				{Status: "VERIFIKASI", Waktu: "30 Nov 2025, 11:00", Petugas: "Admin Kelurahan", Catatan: "Verifikasi dokumen pendukung"},
				{Status: "DAFTAR", Waktu: "30 Nov 2025, 09:30", Petugas: "Sistem", Catatan: "Permohonan berhasil didaftarkan"},
			},
		},
		"4": {
			ID:               "4",
			KodeBooking:      "JKT-02-G7H8",
			NIK:              "3201234567890004",
			NamaLengkap:      "Dewi Lestari",
			TempatLahir:      "Yogyakarta",
			TanggalLahir:     "18 September 1992",
			JenisKelamin:     "Perempuan",
			Alamat:           "Jl. Menteng Dalam No. 56",
			RT:               "007",
			RW:               "010",
			Kelurahan:        "Menteng",
			Kecamatan:        "Menteng",
			Agama:            "Hindu",
			StatusPerkawinan: "Belum Kawin",
			Pekerjaan:        "PNS",
			Kewarganegaraan:  "WNI",
			NoTelp:           "084567890123",
			JenisPermohonan:  "RUSAK",
			AlasanPermohonan: "KTP rusak terkena air, tulisan tidak terbaca.",
			StatusTerkini:    "SELESAI",
			TanggalDaftar:    "29 November 2025",
			JadwalSesi:       "Jumat, 29 Nov 2025 - 13:00 WIB",
			NomorAntrian:     15,
			Catatan:          "",
			RiwayatStatus: []RiwayatStatusItem{
				{Status: "SELESAI", Waktu: "29 Nov 2025, 16:30", Petugas: "Admin Kecamatan", Catatan: "KTP sudah diambil oleh pemohon"},
				{Status: "SIAP_AMBIL", Waktu: "29 Nov 2025, 15:00", Petugas: "Admin Kecamatan", Catatan: "KTP sudah siap diambil"},
				{Status: "PROSES", Waktu: "29 Nov 2025, 14:00", Petugas: "Budi Santoso", Catatan: "Proses pencetakan"},
				{Status: "VERIFIKASI", Waktu: "29 Nov 2025, 13:30", Petugas: "Admin Kelurahan", Catatan: "Verifikasi KTP rusak"},
				{Status: "DAFTAR", Waktu: "29 Nov 2025, 13:00", Petugas: "Sistem", Catatan: "Permohonan berhasil didaftarkan"},
			},
		},
		"5": {
			ID:               "5",
			KodeBooking:      "JKT-02-I9J0",
			NIK:              "3201234567890005",
			NamaLengkap:      "Rudi Hermawan",
			TempatLahir:      "Semarang",
			TanggalLahir:     "3 Maret 1988",
			JenisKelamin:     "Laki-laki",
			Alamat:           "Jl. Pegangsaan Timur No. 23",
			RT:               "002",
			RW:               "006",
			Kelurahan:        "Pegangsaan",
			Kecamatan:        "Menteng",
			Agama:            "Islam",
			StatusPerkawinan: "Cerai Hidup",
			Pekerjaan:        "Pedagang",
			Kewarganegaraan:  "WNI",
			NoTelp:           "085678901234",
			JenisPermohonan:  "BARU",
			AlasanPermohonan: "Pembuatan KTP baru.",
			StatusTerkini:    "DITOLAK",
			TanggalDaftar:    "28 November 2025",
			JadwalSesi:       "Kamis, 28 Nov 2025 - 08:00 WIB",
			NomorAntrian:     5,
			Catatan:          "Dokumen KK tidak sesuai dengan data yang diinput. Harap perbaiki dan ajukan ulang.",
			RiwayatStatus: []RiwayatStatusItem{
				{Status: "DITOLAK", Waktu: "28 Nov 2025, 10:00", Petugas: "Admin Kelurahan", Catatan: "Data KK tidak sesuai dengan form pendaftaran"},
				{Status: "VERIFIKASI", Waktu: "28 Nov 2025, 09:00", Petugas: "Admin Kelurahan", Catatan: "Memverifikasi kelengkapan dokumen"},
				{Status: "DAFTAR", Waktu: "28 Nov 2025, 08:00", Petugas: "Sistem", Catatan: "Permohonan berhasil didaftarkan"},
			},
		},
	}

	detail, exists := detailMap[permohonanID]
	if !exists {
		// Return a not found message
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<div class="text-center py-8 text-slate-500"><p>Permohonan tidak ditemukan</p></div>`))
		return
	}

	PermohonanDetailContent(detail).Render(r.Context(), w)
}

func (h *Handler) JadwalHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/petugas/login", http.StatusSeeOther)
		return
	}

	// Mock data for now - will be replaced with actual DB queries
	data := JadwalPageData{
		UserName:    user.UserName,
		UserRole:    formatRole(user.UserRole),
		ActivePage:  "jadwal",
		CurrentWeek: "2 - 8 Des 2025",
		Kelurahan: []KelurahanOption{
			{ID: 1, Nama: "Kelurahan Menteng"},
			{ID: 2, Nama: "Kelurahan Cikini"},
			{ID: 3, Nama: "Kelurahan Gondangdia"},
		},
		List: []JadwalItem{
			{ID: "1", Tanggal: "2025-12-02", TanggalFormat: "Senin, 2 Des", JamMulai: "08:00", JamSelesai: "11:00", NamaKelurahan: "Kelurahan Menteng", KuotaTerisi: 32, KuotaMaksimal: 50, StatusSesi: "BUKA"},
			{ID: "2", Tanggal: "2025-12-02", TanggalFormat: "Senin, 2 Des", JamMulai: "13:00", JamSelesai: "15:00", NamaKelurahan: "Kelurahan Menteng", KuotaTerisi: 50, KuotaMaksimal: 50, StatusSesi: "PENUH"},
			{ID: "3", Tanggal: "2025-12-03", TanggalFormat: "Selasa, 3 Des", JamMulai: "08:00", JamSelesai: "11:00", NamaKelurahan: "Kelurahan Cikini", KuotaTerisi: 15, KuotaMaksimal: 50, StatusSesi: "BUKA"},
			{ID: "4", Tanggal: "2025-12-03", TanggalFormat: "Selasa, 3 Des", JamMulai: "13:00", JamSelesai: "16:00", NamaKelurahan: "Kelurahan Gondangdia", KuotaTerisi: 0, KuotaMaksimal: 40, StatusSesi: "BUKA"},
			{ID: "5", Tanggal: "2025-12-04", TanggalFormat: "Rabu, 4 Des", JamMulai: "08:00", JamSelesai: "12:00", NamaKelurahan: "Kelurahan Menteng", KuotaTerisi: 28, KuotaMaksimal: 50, StatusSesi: "ISTIRAHAT"},
		},
	}

	JadwalPage(data).Render(r.Context(), w)
}

func formatRole(role string) string {
	switch role {
	case session.RoleAdminKecamatan:
		return "Admin Kecamatan"
	case session.RoleAdminKelurahan:
		return "Admin Kelurahan"
	default:
		return role
	}
}

// PermohonanStatusFormHandler returns the status update form partial
func (h *Handler) PermohonanStatusFormHandler(w http.ResponseWriter, r *http.Request) {
	permohonanID := chi.URLParam(r, "id")

	// Get current status from mock data (will be replaced with DB query)
	currentStatus, exists := mockPermohonanStatus[permohonanID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<div class="text-center py-8 text-red-500"><p>Permohonan tidak ditemukan</p></div>`))
		return
	}

	StatusUpdateForm(permohonanID, currentStatus).Render(r.Context(), w)
}

// UpdateStatusHandler handles the status update submission
func (h *Handler) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	permohonanID := r.FormValue("id")
	newStatus := r.FormValue("status")
	catatan := r.FormValue("catatan")

	// Validate status
	validStatuses := map[string]bool{
		"VERIFIKASI": true,
		"PROSES":     true,
		"SIAP_AMBIL": true,
		"SELESAI":    true,
		"DITOLAK":    true,
	}
	if !validStatuses[newStatus] {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="text-red-500">Status tidak valid</div>`))
		return
	}

	// Update mock data (will be replaced with DB update)
	if _, exists := mockPermohonanStatus[permohonanID]; !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<div class="text-red-500">Permohonan tidak ditemukan</div>`))
		return
	}
	mockPermohonanStatus[permohonanID] = newStatus

	// Log catatan for now (will be saved to riwayat_status in DB)
	_ = catatan

	// Return success response with HX-Trigger to close dialog and refresh data
	w.Header().Set("HX-Trigger", `{"closeDialog": "status-dialog", "refreshPermohonan": true}`)
	w.Header().Set("HX-Reswap", "none")
	w.WriteHeader(http.StatusOK)
}
