package user

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nobuww/simpel-ktp/internal/features/common"
	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
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
		// If error, use empty stats
		stats.Total = 0
		stats.Verifikasi = 0
		stats.Proses = 0
		stats.SiapAmbil = 0
		stats.Selesai = 0
		stats.Ditolak = 0
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
			NomorAntrian:    int(p.NomorAntrian.Int16),
		}

		if p.JadwalTanggal.Valid {
			item.JadwalTanggal = p.JadwalTanggal.Time.Format("02 Jan 2006")
		}
		if p.JadwalJamMulai.Valid {
			// pgtype.Time stores microseconds since midnight
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
