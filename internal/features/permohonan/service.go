package permohonan

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
)

type Service interface {
	GetFormData(ctx context.Context, nik string) (FormData, error)
	GetAvailableJadwal(ctx context.Context, kelurahanID *int32) ([]JadwalOption, error)
	GetLocations(ctx context.Context) ([]LocationOption, error)
	CreatePermohonan(ctx context.Context, req CreatePermohonanRequest) (uuid.UUID, error)
	GetSuccessData(ctx context.Context, permohonanID string, applicationType string) (SuccessData, error)
}

type CreatePermohonanRequest struct {
	UserID    string
	JadwalID  string
	Type      string
	Documents []DocumentFile
}

type DocumentFile struct {
	Type string // e.g., "KK", "KTP", "SURAT_POLISI"
	Path string
}

type PermohonanService struct {
	repo store.Repository
}

func NewService(repo store.Repository) *PermohonanService {
	return &PermohonanService{
		repo: repo,
	}
}

func (s *PermohonanService) GetFormData(ctx context.Context, nik string) (FormData, error) {
	formData := FormData{
		NIK:    nik,
		Errors: make(map[string]string),
	}

	profile, err := s.repo.GetPendudukProfile(ctx, nik)
	if err == nil {
		formData.NamaLengkap = profile.NamaLengkap
		formData.JenisKelamin = profile.JenisKelamin
		formData.Alamat = profile.Alamat.String
		formData.NoHP = profile.NoHp.String
		formData.Email = profile.Email.String
		formData.NamaKelurahan = profile.NamaKelurahan.String
	} else {
		return formData, err
	}

	return formData, nil
}

func (s *PermohonanService) GetAvailableJadwal(ctx context.Context, kelurahanID *int32) ([]JadwalOption, error) {
	today := time.Now()
	nextMonth := today.AddDate(0, 1, 0)

	todayPg := pgtype.Date{Time: today, Valid: true}
	nextMonthPg := pgtype.Date{Time: nextMonth, Valid: true}

	var targetKecamatan bool
	var kelurahanIDPg pgtype.Int4
	if kelurahanID != nil && *kelurahanID == 0 {
		targetKecamatan = true
		// Pass Valid: false to fetch all, will filter in loop
		kelurahanIDPg = pgtype.Int4{Valid: false}
	} else if kelurahanID != nil {
		kelurahanIDPg = pgtype.Int4{Int32: *kelurahanID, Valid: true}
	} else {
		kelurahanIDPg = pgtype.Int4{Valid: false}
	}

	jadwalList, err := s.repo.ListJadwalSesi(ctx, pg_store.ListJadwalSesiParams{
		Tanggal:     todayPg,
		Tanggal_2:   nextMonthPg,
		KelurahanID: kelurahanIDPg,
	})
	if err != nil {
		return nil, err
	}

	options := make([]JadwalOption, 0, len(jadwalList))
	for _, j := range jadwalList {
		if targetKecamatan && j.NamaKelurahan != "Kecamatan Pademangan" {
			continue
		}
		if j.StatusSesi.String != "BUKA" {
			continue
		}
		kuotaSisa := int(j.KuotaMaksimal - j.KuotaTerisi)
		if kuotaSisa <= 0 {
			continue
		}

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

	return options, nil
}

func (s *PermohonanService) GetLocations(ctx context.Context) ([]LocationOption, error) {
	locations, err := s.repo.ListAllKelurahan(ctx)
	if err != nil {
		return nil, err
	}

	options := make([]LocationOption, 0, len(locations)+1)
	options = append(options, LocationOption{
		ID:    0,
		Label: "Kecamatan Pademangan",
	})

	for _, l := range locations {
		options = append(options, LocationOption{
			ID:    l.ID,
			Label: l.NamaKelurahan,
		})
	}

	return options, nil
}

func (s *PermohonanService) CreatePermohonan(ctx context.Context, req CreatePermohonanRequest) (uuid.UUID, error) {
	jadwalUUID, err := uuid.Parse(req.JadwalID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid jadwal ID: %w", err)
	}

	nikText := pgtype.Text{String: req.UserID, Valid: true}

	// Check for active permohonan
	counts, err := s.repo.CountPermohonanByNIK(ctx, nikText)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check existing applications: %w", err)
	}

	activeCount := counts.Verifikasi + counts.Proses + counts.SiapAmbil
	if activeCount > 0 {
		return uuid.Nil, fmt.Errorf("anda masih memiliki permohonan yang sedang berjalan (Status: Verifikasi/Proses/Siap Ambil)")
	}

	jadwalUUIDPg := pgtype.UUID{Bytes: jadwalUUID, Valid: true}

	jenisPermohonan := ""
	switch req.Type {
	case "baru":
		jenisPermohonan = "BARU"
	case "hilang":
		jenisPermohonan = "HILANG"
	case "rusak":
		jenisPermohonan = "RUSAK"
	case "ubah":
		jenisPermohonan = "UPDATE"
	default:
		jenisPermohonan = req.Type
	}

	permohonanID, err := s.repo.CreatePermohonan(ctx, pg_store.CreatePermohonanParams{
		Nik:             nikText,
		JadwalSesiID:    jadwalUUIDPg,
		JenisPermohonan: jenisPermohonan,
	})
	if err != nil {
		return uuid.Nil, err
	}

	permohonanUUID := pgtype.UUID{Bytes: permohonanID, Valid: true}
	for _, doc := range req.Documents {
		err = s.repo.CreateDokumenSyarat(ctx, pg_store.CreateDokumenSyaratParams{
			PermohonanID: permohonanUUID,
			FilePath:     doc.Path,
			JenisDokumen: doc.Type,
		})
		if err != nil {
			return permohonanID, fmt.Errorf("failed to save document %s: %w", doc.Type, err)
		}
	}

	return permohonanID, nil
}

func (s *PermohonanService) GetSuccessData(ctx context.Context, permohonanID string, applicationType string) (SuccessData, error) {
	permohonanUUID, err := uuid.Parse(permohonanID)
	if err != nil {
		return SuccessData{}, err
	}

	detail, err := s.repo.GetPermohonanDetail(ctx, permohonanUUID)
	if err != nil {
		return SuccessData{}, err
	}

	successData := SuccessData{
		PermohonanID:    permohonanID,
		KodeBooking:     detail.KodeBooking.String,
		ApplicationType: formatApplicationType(applicationType),
		JadwalTanggal:   formatDate(detail.JadwalTanggal),
		JadwalJam:       formatTime(detail.JadwalJamMulai) + " - " + formatTime(detail.JadwalJamSelesai),
		NamaKelurahan:   detail.NamaKelurahan.String,
	}

	return successData, nil
}

// Helpers

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
