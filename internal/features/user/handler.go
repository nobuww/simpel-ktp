package user

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nobuww/simpel-ktp/internal/features/common"
	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
	"github.com/nobuww/simpel-ktp/ui/components"
)

// Handler manages user-related HTTP handlers
type Handler struct {
	store *store.Store
}

// New creates a new user handler with the required dependencies
func New(s *store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}
	ctx := r.Context()

	// Convert NIK to pgtype.Text
	nikText := pgtype.Text{String: user.UserID, Valid: true}

	// Get user's permohonan stats
	stats, err := h.store.CountPermohonanByNIK(ctx, nikText)
	if err != nil {
		stats = pg_store.CountPermohonanByNIKRow{}
	}

	// Get recent permohonan (limit 5)
	permohonanList, err := h.store.GetPermohonanByNIK(ctx, pg_store.GetPermohonanByNIKParams{
		Nik:    nikText,
		Limit:  5,
		Offset: 0,
	})
	if err != nil {
		permohonanList = nil
	}

	data := DashboardData{
		UserName: user.UserName,
		NIK:      user.UserID,
		Stats: UserStats{
			Total:      stats.Total,
			Verifikasi: stats.Verifikasi,
			Proses:     stats.Proses,
			SiapAmbil:  stats.SiapAmbil,
			Selesai:    stats.Selesai,
			Ditolak:    stats.Ditolak,
		},
		RecentPermohonan: convertPermohonanList(permohonanList),
	}

	DashboardPage(data).Render(ctx, w)
}

func convertPermohonanList(list []pg_store.GetPermohonanByNIKRow) []PermohonanItem {
	items := make([]PermohonanItem, 0, len(list))
	for _, p := range list {
		item := PermohonanItem{
			ID:              p.ID.String(),
			KodeBooking:     p.KodeBooking.String,
			JenisPermohonan: p.JenisPermohonan,
			StatusTerkini:   p.StatusTerkini.String,
		}
		if p.NomorAntrian.Valid {
			item.NomorAntrian = int(p.NomorAntrian.Int16)
		}

		if p.JadwalTanggal.Valid {
			item.JadwalTanggal = p.JadwalTanggal.Time.Format("02 Jan 2006")
		}
		if p.JadwalJamMulai.Valid {
			hour := p.JadwalJamMulai.Microseconds / 3600000000
			minute := (p.JadwalJamMulai.Microseconds % 3600000000) / 60000000
			item.JadwalJam = fmt.Sprintf("%02d:%02d", hour, minute)
		}
		if p.LokasiKelurahan.Valid {
			item.LokasiKelurahan = p.LokasiKelurahan.String
		}
		if p.CreatedAt.Valid {
			item.TanggalDaftar = p.CreatedAt.Time.Format("02 Jan 2006")
		}

		items = append(items, item)
	}
	return items
}

// StatusDetailHandler handles the status tracking detail page
func (h *Handler) StatusDetailHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := common.GetUserOrRedirect(w, r, "/login")
	if !ok {
		return
	}

	kode := r.URL.Query().Get("kode")
	ctx := r.Context()

	var detail pg_store.GetPermohonanByKodeBookingRow
	var err error

	if kode != "" {
		detail, err = h.store.GetPermohonanByKodeBooking(ctx, pgtype.Text{String: kode, Valid: true})
		if err != nil {
			common.WriteNotFound(w, "Permohonan tidak ditemukan")
			return
		}
	} else {
		// Fallback to latest
		nikText := pgtype.Text{String: user.UserID, Valid: true}
		list, errList := h.store.GetPermohonanByNIK(ctx, pg_store.GetPermohonanByNIKParams{
			Nik:    nikText,
			Limit:  1,
			Offset: 0,
		})
		if errList == nil && len(list) > 0 {
			if list[0].KodeBooking.Valid {
				kode = list[0].KodeBooking.String
				detail, err = h.store.GetPermohonanByKodeBooking(ctx, pgtype.Text{String: kode, Valid: true})
				if err != nil {
					common.WriteNotFound(w, "Permohonan tidak ditemukan")
					return
				}
			}
		} else {
			// No permohonan found
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	stages, currentIdx := calculateStages(detail)

	data := StatusDetailData{
		UserName:        user.UserName,
		KodeBooking:     detail.KodeBooking.String,
		JenisPermohonan: detail.JenisPermohonan,
		Stages:          stages,
		CurrentStageIdx: currentIdx,
		NextSteps:       determineNextSteps(detail),
	}
	if detail.TanggalDaftar.Valid {
		data.TanggalDaftar = detail.TanggalDaftar.Time.Format("02 Jan 2006")
	}

	StatusDetailPage(data).Render(ctx, w)
}

func calculateStages(p pg_store.GetPermohonanByKodeBookingRow) ([]components.TrackerStage, int) {
	stages := []components.TrackerStage{
		{ID: "1", Name: "Pengajuan", Description: "Permohonan diterima", Status: components.StatusPending},
		{ID: "2", Name: "Verifikasi", Description: "Verifikasi dokumen", Status: components.StatusPending},
		{ID: "3", Name: "Proses", Description: "Pencetakan KTP", Status: components.StatusPending},
		{ID: "4", Name: "Siap Ambil", Description: "Siap diambil di kelurahan", Status: components.StatusPending},
		{ID: "5", Name: "Selesai", Description: "Selesai", Status: components.StatusPending},
	}

	if p.TanggalDaftar.Valid {
		stages[0].StartDate = p.TanggalDaftar.Time.Format("02 Jan 2006, 15:04")
	}

	status := ""
	if p.StatusTerkini.Valid {
		status = p.StatusTerkini.String
	}

	idx := 0
	switch status {
	case "VERIFIKASI":
		idx = 1
		stages[0].Status = components.StatusCompleted
		stages[1].Status = components.StatusInProgress
	case "PROSES":
		idx = 2
		stages[0].Status = components.StatusCompleted
		stages[1].Status = components.StatusCompleted
		stages[2].Status = components.StatusInProgress
	case "SIAP_AMBIL":
		idx = 3
		stages[0].Status = components.StatusCompleted
		stages[1].Status = components.StatusCompleted
		stages[2].Status = components.StatusCompleted
		stages[3].Status = components.StatusInProgress
	case "SELESAI":
		idx = 4
		stages[0].Status = components.StatusCompleted
		stages[1].Status = components.StatusCompleted
		stages[2].Status = components.StatusCompleted
		stages[3].Status = components.StatusCompleted
		stages[4].Status = components.StatusCompleted
	case "DITOLAK":
		idx = 1
		stages[0].Status = components.StatusCompleted
		stages[1].Status = components.StatusBlocked
		stages[1].Notes = "Permohonan ditolak."
	}

	return stages, idx
}

func determineNextSteps(p pg_store.GetPermohonanByKodeBookingRow) []NextStep {
	status := ""
	if p.StatusTerkini.Valid {
		status = p.StatusTerkini.String
	}

	steps := []NextStep{}

	switch status {
	case "VERIFIKASI":
		steps = append(steps, NextStep{
			Title:       "Tunggu Verifikasi",
			Description: "Mohon tunggu petugas memverifikasi dokumen Anda.",
			IsPrimary:   false,
		})
	case "SIAP_AMBIL":
		steps = append(steps, NextStep{
			Title:       "Ambil KTP",
			Description: "Silakan datang ke kelurahan untuk mengambil KTP.",
			IsPrimary:   true,
			ActionURL:   "#",
		})
	case "DITOLAK":
		steps = append(steps, NextStep{
			Title:       "Ajukan Ulang",
			Description: "Silakan perbaiki dokumen dan ajukan ulang.",
			IsPrimary:   true,
			ActionURL:   "/permohonan/baru", // Generic link
		})
	}

	return steps
}
