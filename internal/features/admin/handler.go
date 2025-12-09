package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nobuww/simpel-ktp/internal/features/common"
	"github.com/nobuww/simpel-ktp/internal/middleware"
	"github.com/nobuww/simpel-ktp/internal/session"
	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
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

func getKelurahanID(user *session.UserSession) pgtype.Int2 {
	if user == nil || user.KelurahanID == nil {
		return pgtype.Int2{Valid: false}
	}
	return pgtype.Int2{Int16: *user.KelurahanID, Valid: true}
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	ctx := r.Context()

	scopeID := getKelurahanID(user)

	// Get Stats
	statsRow, err := h.store.GetAdminDashboardStats(ctx, scopeID)
	if err != nil {
		statsRow = pg_store.GetAdminDashboardStatsRow{}
	}

	stats := PermohonanStats{
		Total:      int(statsRow.TotalPermohonan),
		Verifikasi: int(statsRow.Verifikasi),
		Proses:     int(statsRow.Proses),
		SiapAmbil:  int(statsRow.SiapAmbil),
		Selesai:    int(statsRow.Selesai),
		Ditolak:    int(statsRow.Ditolak),
	}

	// Get Recent Permohonan (Limit 5)
	recentRows, err := h.store.ListPermohonanAdmin(ctx, pg_store.ListPermohonanAdminParams{
		Limit:       5,
		Offset:      0,
		Search:      pgtype.Text{Valid: false},
		Status:      pgtype.Text{Valid: false},
		KelurahanID: scopeID,
	})
	if err != nil {
		recentRows = nil
	}
	recent := convertPermohonanList(recentRows)

	// Get Today's Jadwal
	todayJadwalRows, err := h.store.ListTodayJadwal(ctx, scopeID)
	if err != nil {
		todayJadwalRows = nil
	}
	todayJadwal := convertTodayJadwalList(todayJadwalRows)

	data := DashboardData{
		UserName:    user.UserName,
		UserRole:    common.FormatRole(user.UserRole),
		ActivePage:  "dashboard",
		Stats:       stats,
		Recent:      recent,
		TodayJadwal: todayJadwal,
	}

	DashboardPage(data).Render(ctx, w)
}

func (h *Handler) PermohonanHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	ctx := r.Context()

	// Pagination
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	// Filters
	search := r.URL.Query().Get("search")
	statusFilter := r.URL.Query().Get("status")

	// Stats for top cards
	scopeID := getKelurahanID(user)

	// Stats for top cards
	statsRow, err := h.store.GetAdminDashboardStats(ctx, scopeID)
	if err != nil {
		statsRow = pg_store.GetAdminDashboardStatsRow{}
	}
	stats := PermohonanStats{
		Total:      int(statsRow.TotalPermohonan),
		Verifikasi: int(statsRow.Verifikasi),
		Proses:     int(statsRow.Proses),
		SiapAmbil:  int(statsRow.SiapAmbil),
		Selesai:    int(statsRow.Selesai),
		Ditolak:    int(statsRow.Ditolak),
	}

	// List
	params := pg_store.ListPermohonanAdminParams{
		Limit:       int32(limit),
		Offset:      int32(offset),
		Search:      pgtype.Text{String: search, Valid: search != ""},
		Status:      pgtype.Text{String: statusFilter, Valid: statusFilter != ""},
		KelurahanID: scopeID,
	}

	listRows, err := h.store.ListPermohonanAdmin(ctx, params)
	if err != nil {
		listRows = nil
	}
	list := convertPermohonanList(listRows)

	data := PermohonanPageData{
		UserName:   user.UserName,
		UserRole:   common.FormatRole(user.UserRole),
		ActivePage: "permohonan",
		Stats:      stats,
		List:       list,
	}

	PermohonanPage(data).Render(ctx, w)
}

func convertPermohonanList(list []pg_store.ListPermohonanAdminRow) []PermohonanItem {
	items := make([]PermohonanItem, 0, len(list))
	for _, p := range list {
		item := PermohonanItem{
			ID:              p.ID.String(),
			KodeBooking:     p.KodeBooking.String,
			NamaLengkap:     p.NamaLengkap,
			JenisPermohonan: p.JenisPermohonan,
			StatusTerkini:   p.StatusTerkini.String,
		}

		if p.Nik.Valid {
			item.NIK = p.Nik.String
		}
		if p.TanggalDaftar.Valid {
			item.TanggalDaftar = p.TanggalDaftar.Time.Format("2006-01-02")
		}

		if p.JadwalTanggal.Valid && p.JadwalJamMulai.Valid {
			dateStr := p.JadwalTanggal.Time.Format("Mon, 2 Jan")
			// Calculate time from microseconds
			micros := p.JadwalJamMulai.Microseconds
			hours := micros / 3600000000
			minutes := (micros % 3600000000) / 60000000
			timeStr := fmt.Sprintf("%02d:%02d", hours, minutes)
			item.JadwalSesi = fmt.Sprintf("%s %s", dateStr, timeStr)
		} else {
			item.JadwalSesi = "-"
		}

		if p.NomorAntrian.Valid {
			item.NomorAntrian = int(p.NomorAntrian.Int16)
		}

		items = append(items, item)
	}
	return items
}

func convertTodayJadwalList(rows []pg_store.ListTodayJadwalRow) []TodayJadwalItem {
	items := make([]TodayJadwalItem, len(rows))
	for i, r := range rows {
		jamMulai := convertMicrosToTime(r.JamMulai.Microseconds)
		jamSelesai := convertMicrosToTime(r.JamSelesai.Microseconds)

		items[i] = TodayJadwalItem{
			ID:            r.ID.String(),
			SessionName:   r.NamaKelurahan,
			Time:          fmt.Sprintf("%s - %s", jamMulai, jamSelesai),
			KuotaTerisi:   int(r.KuotaTerisi),
			KuotaMaksimal: int(r.KuotaMaksimal),
			StatusSesi:    r.StatusSesi.String,
		}
	}
	return items
}

func (h *Handler) PermohonanDetailHandler(w http.ResponseWriter, r *http.Request) {
	permohonanIDStr := chi.URLParam(r, "id")
	permohonanID, err := uuid.Parse(permohonanIDStr)
	if err != nil {
		common.WriteNotFound(w, "ID Permohonan tidak valid")
		return
	}

	ctx := r.Context()

	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	scopeID := getKelurahanID(user)

	detailRow, err := h.store.GetPermohonanDetailAdmin(ctx, pg_store.GetPermohonanDetailAdminParams{
		ID:          permohonanID,
		KelurahanID: scopeID,
	})
	if err != nil {
		common.WriteNotFound(w, "Permohonan tidak ditemukan")
		return
	}

	historyRows, err := h.store.GetRiwayatStatusByPermohonan(ctx, pgtype.UUID{Bytes: permohonanID, Valid: true})
	if err != nil {
		// Just treat as empty
		historyRows = []pg_store.GetRiwayatStatusByPermohonanRow{}
	}

	// Fetch documents
	dokumenRows, err := h.store.GetDokumenByPermohonan(ctx, pgtype.UUID{Bytes: permohonanID, Valid: true})
	if err != nil {
		dokumenRows = []pg_store.GetDokumenByPermohonanRow{}
	}

	detail := PermohonanDetail{
		ID:              permohonanIDStr,
		KodeBooking:     detailRow.KodeBooking.String,
		NIK:             detailRow.Nik.String,
		NamaLengkap:     detailRow.NamaLengkap,
		JenisKelamin:    detailRow.JenisKelamin,
		Alamat:          detailRow.Alamat.String,
		NoTelp:          detailRow.NoTelp.String,
		JenisPermohonan: detailRow.JenisPermohonan,
		StatusTerkini:   detailRow.StatusTerkini.String,
		Kelurahan:       detailRow.Kelurahan.String,
		RiwayatStatus:   convertHistoryList(historyRows),
		Dokumen:         convertDokumenList(dokumenRows),
	}

	if detailRow.TanggalDaftar.Valid {
		detail.TanggalDaftar = detailRow.TanggalDaftar.Time.Format("2 Jan 2006")
	}
	if detailRow.JadwalTanggal.Valid && detailRow.JadwalJamMulai.Valid {
		d := detailRow.JadwalTanggal.Time.Format("Mon, 2 Jan 2006")
		micros := detailRow.JadwalJamMulai.Microseconds
		h := micros / 3600000000
		m := (micros % 3600000000) / 60000000
		detail.JadwalSesi = fmt.Sprintf("%s - %02d:%02d WIB", d, h, m)
	}
	if detailRow.NomorAntrian.Valid {
		detail.NomorAntrian = int(detailRow.NomorAntrian.Int16)
	}

	PermohonanDetailContent(detail).Render(ctx, w)
}

func convertDokumenList(rows []pg_store.GetDokumenByPermohonanRow) []DokumenItem {
	items := make([]DokumenItem, len(rows))
	for i, r := range rows {
		item := DokumenItem{
			ID:           r.ID.String(),
			JenisDokumen: r.JenisDokumen,
			FilePath:     r.FilePath,
		}
		if r.UploadedAt.Valid {
			item.UploadedAt = r.UploadedAt.Time.Format("2 Jan 2006, 15:04")
		}
		items[i] = item
	}
	return items
}

func convertHistoryList(rows []pg_store.GetRiwayatStatusByPermohonanRow) []RiwayatStatusItem {
	items := make([]RiwayatStatusItem, len(rows))
	for i, r := range rows {
		item := RiwayatStatusItem{
			Status:  r.StatusBaru,
			Catatan: r.CatatanProses.String,
		}
		if r.WaktuProses.Valid {
			item.Waktu = r.WaktuProses.Time.Format("2 Jan 2006, 15:04")
		}
		if r.NamaPetugas.Valid {
			item.Petugas = r.NamaPetugas.String
		} else {
			item.Petugas = "Sistem"
		}
		items[i] = item
	}
	return items
}

func (h *Handler) JadwalHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	ctx := r.Context()
	scopeID := getKelurahanID(user)

	// Determine date range
	refDateStr := r.URL.Query().Get("ref_date")
	refDate := time.Now()
	if refDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", refDateStr); err == nil {
			refDate = parsed
		}
	}

	weekAction := r.URL.Query().Get("week")
	startOfWeek := getStartOfWeek(refDate)
	switch weekAction {
	case "prev":
		startOfWeek = startOfWeek.AddDate(0, 0, -7)
	case "next":
		startOfWeek = startOfWeek.AddDate(0, 0, 7)
	}
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	// Determine if admin is kecamatan level
	isKecamatan := !scopeID.Valid

	// Get kelurahan name for display (for kelurahan admin)
	kelurahanName := ""
	if scopeID.Valid {
		kelRow, err := h.store.GetKelurahanById(ctx, scopeID.Int16)
		if err == nil {
			kelurahanName = kelRow.NamaKelurahan
		}
	}

	// Filter by user's kelurahan (kecamatan sees only kecamatan schedules - NULL kelurahan_id)
	var filterKelurahanID pgtype.Int4
	if scopeID.Valid {
		filterKelurahanID = pgtype.Int4{Int32: int32(scopeID.Int16), Valid: true}
	}

	// Fetch Jadwal
	listParams := pg_store.ListJadwalSesiParams{
		Tanggal:     pgtype.Date{Time: startOfWeek, Valid: true},
		Tanggal_2:   pgtype.Date{Time: endOfWeek, Valid: true},
		KelurahanID: filterKelurahanID,
	}

	rows, err := h.store.ListJadwalSesi(ctx, listParams)
	if err != nil {
		rows = []pg_store.ListJadwalSesiRow{}
	}

	items := make([]JadwalItem, len(rows))
	for i, r := range rows {
		items[i] = JadwalItem{
			ID:            r.ID.String(),
			Tanggal:       r.Tanggal.Time.Format("2006-01-02"),
			TanggalFormat: r.Tanggal.Time.Format("Mon, 2 Jan 2006"),
			JamMulai:      convertMicrosToTime(r.JamMulai.Microseconds),
			JamSelesai:    convertMicrosToTime(r.JamSelesai.Microseconds),
			NamaKelurahan: r.NamaKelurahan,
			KuotaTerisi:   int(r.KuotaTerisi),
			KuotaMaksimal: int(r.KuotaMaksimal),
			StatusSesi:    r.StatusSesi.String,
		}
	}

	data := JadwalPageData{
		UserName:        user.UserName,
		UserRole:        common.FormatRole(user.UserRole),
		ActivePage:      "jadwal",
		CurrentWeek:     fmt.Sprintf("%s - %s", startOfWeek.Format("2 Jan"), endOfWeek.Format("2 Jan 2006")),
		StartOfWeekDate: startOfWeek.Format("2006-01-02"),
		KelurahanName:   kelurahanName,
		IsKecamatan:     isKecamatan,
		List:            items,
	}

	// Determine if we should render partials (week navigation) or full page
	// We only render partials if the request targets the content area specifically
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "jadwal-content" {
		JadwalCardView(data.List).Render(ctx, w)
		WeekNavigation(data, true).Render(ctx, w)
		return
	}

	JadwalPage(data).Render(r.Context(), w)
}

func (h *Handler) CreateJadwalHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.WriteError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	ctx := r.Context()
	scopeID := getKelurahanID(user)

	// Admin Kelurahan: use their kelurahan ID
	// Admin Kecamatan: use NULL (for Kantor Kecamatan)
	var lokasiKelurahanID pgtype.Int2
	if scopeID.Valid {
		lokasiKelurahanID = scopeID
	} else {
		// Admin Kecamatan - lokasi is Kantor Kecamatan (NULL)
		lokasiKelurahanID = pgtype.Int2{Valid: false}
	}

	tanggalStr := r.FormValue("tanggal")
	jamMulaiStr := r.FormValue("jam_mulai")
	jamSelesaiStr := r.FormValue("jam_selesai")
	kuotaStr := r.FormValue("kuota_maksimal")

	// Parse inputs
	tanggal, err := time.Parse("2006-01-02", tanggalStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "Format tanggal salah")
		return
	}

	jamMulai, err := parseTime(jamMulaiStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "Format jam mulai salah")
		return
	}

	jamSelesai, err := parseTime(jamSelesaiStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "Format jam selesai salah")
		return
	}

	kuota, _ := strconv.Atoi(kuotaStr)
	if kuota <= 0 {
		kuota = 50 // Default quota
	}

	id, err := h.store.CreateJadwalSesi(ctx, pg_store.CreateJadwalSesiParams{
		LokasiKelurahanID: lokasiKelurahanID,
		Tanggal:           pgtype.Date{Time: tanggal, Valid: true},
		JamMulai:          pgtype.Time{Microseconds: jamMulai, Valid: true},
		JamSelesai:        pgtype.Time{Microseconds: jamSelesai, Valid: true},
		KuotaMaksimal:     int16(kuota),
	})
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, "Gagal membuat jadwal: "+err.Error())
		return
	}
	_ = id

	common.HXTrigger(w, `{"closeDialog": "create-jadwal-dialog", "refreshJadwal": true}`)
	common.HXRedirect(w, "/admin/jadwal")
}

func parseTime(t string) (int64, error) {
	parsed, err := time.Parse("15:04", t)
	if err != nil {
		return 0, err
	}
	// Microseconds since midnight
	return int64(parsed.Hour()*3600+parsed.Minute()*60) * 1000000, nil
}

func (h *Handler) GenerateJadwalHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	ctx := r.Context()
	scopeID := getKelurahanID(user)

	// Admin Kelurahan: use their kelurahan ID
	// Admin Kecamatan: use NULL (for Kantor Kecamatan)
	var lokasiKelurahanID pgtype.Int2
	if scopeID.Valid {
		lokasiKelurahanID = scopeID
	} else {
		// Admin Kecamatan - lokasi is Kantor Kecamatan (NULL)
		lokasiKelurahanID = pgtype.Int2{Valid: false}
	}

	startDate := time.Now().AddDate(0, 0, 1) // Start tomorrow
	for i := 0; i < 30; i++ {
		date := startDate.AddDate(0, 0, i)
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}

		pgDate := pgtype.Date{Time: date, Valid: true}

		// Session 1: 09:00 - 12:00
		h.store.CreateJadwalSesi(ctx, pg_store.CreateJadwalSesiParams{
			LokasiKelurahanID: lokasiKelurahanID,
			Tanggal:           pgDate,
			JamMulai:          pgtype.Time{Microseconds: 9 * 3600 * 1000000, Valid: true},
			JamSelesai:        pgtype.Time{Microseconds: 12 * 3600 * 1000000, Valid: true},
			KuotaMaksimal:     50,
		})
		// Session 2: 13:00 - 15:00
		h.store.CreateJadwalSesi(ctx, pg_store.CreateJadwalSesiParams{
			LokasiKelurahanID: lokasiKelurahanID,
			Tanggal:           pgDate,
			JamMulai:          pgtype.Time{Microseconds: 13 * 3600 * 1000000, Valid: true},
			JamSelesai:        pgtype.Time{Microseconds: 15 * 3600 * 1000000, Valid: true},
			KuotaMaksimal:     50,
		})
	}

	common.HXRedirect(w, "/admin/jadwal")
}

func getStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -weekday+1)
}

func (h *Handler) PendudukHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	ctx := r.Context()
	scopeID := getKelurahanID(user)

	// Fetch Stats
	statsRow, err := h.store.GetPendudukStatsAdmin(ctx, scopeID)
	if err != nil {
		statsRow = pg_store.GetPendudukStatsAdminRow{}
	}

	stats := PendudukStats{
		Total:     int(statsRow.Total),
		LakiLaki:  int(statsRow.LakiLaki),
		Perempuan: int(statsRow.Perempuan),
		WajibKTP:  int(statsRow.WajibKtp),
	}

	// Fetch List
	search := r.URL.Query().Get("search")
	listParams := pg_store.ListPendudukAdminParams{
		Search:      pgtype.Text{String: search, Valid: search != ""},
		KelurahanID: scopeID,
	}

	rows, err := h.store.ListPendudukAdmin(ctx, listParams)
	if err != nil {
		rows = []pg_store.ListPendudukAdminRow{}
	}

	list := make([]PendudukItem, len(rows))
	for i, r := range rows {
		list[i] = PendudukItem{
			ID:           r.Nik,
			NIK:          r.Nik,
			NamaLengkap:  r.NamaLengkap,
			JenisKelamin: r.JenisKelamin,
			Alamat:       r.Alamat.String,
			Kelurahan:    r.NamaKelurahan.String,
			Email:        r.Email.String,
			NoHP:         r.NoHp.String,
		}
	}

	data := PendudukPageData{
		UserName:   user.UserName,
		UserRole:   common.FormatRole(user.UserRole),
		ActivePage: "penduduk",
		Stats:      stats,
		List:       list,
	}

	PendudukPage(data).Render(r.Context(), w)
}

// PermohonanStatusFormHandler returns the status update form partial
func (h *Handler) PermohonanStatusFormHandler(w http.ResponseWriter, r *http.Request) {
	permohonanIDStr := chi.URLParam(r, "id")
	permohonanID, err := uuid.Parse(permohonanIDStr)
	if err != nil {
		common.WriteNotFound(w, "ID tidak valid")
		return
	}

	ctx := r.Context()

	// Get current status
	statusRow, err := h.store.GetPermohonanStatusById(ctx, permohonanID)
	if err != nil {
		common.WriteError(w, http.StatusNotFound, "Permohonan tidak ditemukan")
		return
	}

	currentStatus := ""
	if statusRow.StatusTerkini.Valid {
		currentStatus = statusRow.StatusTerkini.String
	}

	StatusUpdateForm(permohonanIDStr, currentStatus).Render(ctx, w)
}

// UpdateStatusHandler handles the status update submission
func (h *Handler) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	permohonanIDStr := r.FormValue("id")
	newStatus := r.FormValue("status")
	catatan := r.FormValue("catatan")

	permohonanID, err := uuid.Parse(permohonanIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	validStatuses := map[string]bool{
		"VERIFIKASI": true,
		"PROSES":     true,
		"SIAP_AMBIL": true,
		"SELESAI":    true,
		"DITOLAK":    true,
	}
	if !validStatuses[newStatus] {
		common.WriteError(w, http.StatusBadRequest, "Status tidak valid")
		return
	}

	ctx := r.Context()

	user := middleware.GetUserFromContext(ctx)
	var petugasID pgtype.UUID
	if user != nil {
		if uid, err := uuid.Parse(user.UserID); err == nil {
			petugasID = pgtype.UUID{Bytes: uid, Valid: true}
		}
	}

	scopeID := getKelurahanID(user)

	err = h.store.ExecTx(ctx, func(q *pg_store.Queries) error {
		if err := q.UpdatePermohonanStatusAdmin(ctx, pg_store.UpdatePermohonanStatusAdminParams{
			ID:            permohonanID,
			StatusTerkini: pgtype.Text{String: newStatus, Valid: true},
			KelurahanID:   scopeID,
		}); err != nil {
			return err
		}

		if err := q.InsertRiwayatStatus(ctx, pg_store.InsertRiwayatStatusParams{
			PermohonanID:  pgtype.UUID{Bytes: permohonanID, Valid: true},
			PetugasID:     petugasID,
			StatusBaru:    newStatus,
			CatatanProses: pgtype.Text{String: catatan, Valid: true},
		}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, "Gagal update status: "+err.Error())
		return
	}

	common.HXTrigger(w, `{"closeDialog": "status-dialog", "refreshPermohonan": true}`)
	common.HXReswap(w, "none")

	w.WriteHeader(http.StatusOK)
}

func convertMicrosToTime(micros int64) string {
	hours := micros / 3600000000
	minutes := (micros % 3600000000) / 60000000
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func (h *Handler) JadwalAntrianHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	user, ok := common.GetUserOrRedirect(w, r, "/petugas/login")
	if !ok {
		return
	}
	ctx := r.Context()

	// Get Jadwal Info
	item, err := h.store.GetJadwalSesiById(ctx, id)
	if err != nil {
		common.WriteNotFound(w, "Jadwal not found")
		return
	}

	jadwalInfo := JadwalItem{
		ID:            item.ID.String(),
		Tanggal:       item.Tanggal.Time.Format("2006-01-02"),
		TanggalFormat: item.Tanggal.Time.Format("Mon, 2 Jan 2006"), // Formatted for display
		JamMulai:      convertMicrosToTime(item.JamMulai.Microseconds),
		JamSelesai:    convertMicrosToTime(item.JamSelesai.Microseconds),
		NamaKelurahan: item.NamaKelurahan,
		KuotaMaksimal: int(item.KuotaMaksimal),
		KuotaTerisi:   int(item.KuotaTerisi),
		StatusSesi:    item.StatusSesi.String,
	}

	// Get Antrian List
	listRows, err := h.store.ListPermohonanByJadwal(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		// handle error or empty
		listRows = []pg_store.ListPermohonanByJadwalRow{}
	}

	antrianList := make([]AntrianItem, len(listRows))
	for i, r := range listRows {
		antrianList[i] = AntrianItem{
			ID:          r.ID.String(),
			KodeBooking: r.KodeBooking.String,
			NIK:         r.Nik.String,
			NamaLengkap: r.NamaLengkap,
			Status:      r.StatusTerkini.String,
			NoAntrian:   int(r.NomorAntrian.Int16),
		}
	}

	data := JadwalAntrianData{
		JadwalInfo:  jadwalInfo,
		AntrianList: antrianList,
		UserName:    user.UserName,
		UserRole:    common.FormatRole(user.UserRole),
		ActivePage:  "jadwal",
	}

	JadwalAntrianPage(data).Render(ctx, w)
}
