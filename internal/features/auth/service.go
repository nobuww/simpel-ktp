package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
)

// Service errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNoPassword  = errors.New("account has no password set")
	ErrNIKExists          = errors.New("NIK already registered")
	ErrEmailExists        = errors.New("email already registered")
	ErrInternalError      = errors.New("internal error")
)

// Service handles authentication business logic
type Service struct {
	store *store.Store
}

// NewService creates a new auth service
func NewService(store *store.Store) *Service {
	return &Service{store: store}
}

// WargaLoginInput contains input for warga login
type WargaLoginInput struct {
	NIK      string
	Password string
}

// WargaLoginResult contains the result of a successful warga login
type WargaLoginResult struct {
	NIK         string
	NamaLengkap string
}

// LoginWarga authenticates a warga (citizen) by NIK and password
func (s *Service) LoginWarga(ctx context.Context, input WargaLoginInput) (*WargaLoginResult, error) {
	penduduk, err := s.store.GetPendudukByNIK(ctx, input.NIK)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, ErrInternalError
	}

	// Check if password is set
	if !penduduk.PasswordHash.Valid || penduduk.PasswordHash.String == "" {
		return nil, ErrAccountNoPassword
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(penduduk.PasswordHash.String), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &WargaLoginResult{
		NIK:         penduduk.Nik,
		NamaLengkap: penduduk.NamaLengkap,
	}, nil
}

// PetugasLoginInput contains input for petugas login
type PetugasLoginInput struct {
	NIP      string
	Password string
}

// PetugasLoginResult contains the result of a successful petugas login
type PetugasLoginResult struct {
	ID          string
	NamaPetugas string
	Role        string
	KelurahanID *int16
}

// LoginPetugas authenticates a petugas (officer) by NIP and password
func (s *Service) LoginPetugas(ctx context.Context, input PetugasLoginInput) (*PetugasLoginResult, error) {
	petugas, err := s.store.GetPetugasByNIP(ctx, pgtype.Text{String: input.NIP, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, ErrInternalError
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(petugas.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	var kelurahanID *int16
	if petugas.KelurahanID.Valid {
		val := petugas.KelurahanID.Int16
		kelurahanID = &val
	}

	return &PetugasLoginResult{
		ID:          petugas.ID.String(),
		NamaPetugas: petugas.NamaPetugas,
		Role:        petugas.Role,
		KelurahanID: kelurahanID,
	}, nil
}

// RegisterInput contains input for warga registration
type RegisterInput struct {
	NIK          string
	NamaLengkap  string
	Email        string
	Password     string
	Alamat       string
	NoHP         string
	JenisKelamin string
	KelurahanID  int16
}

// RegisterResult contains the result of a successful registration
type RegisterResult struct {
	NIK         string
	NamaLengkap string
}

// RegisterWarga creates a new warga (citizen) account
func (s *Service) RegisterWarga(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	// Check if NIK already exists
	exists, err := s.store.CheckPendudukExists(ctx, input.NIK)
	if err != nil {
		return nil, ErrInternalError
	}
	if exists {
		return nil, ErrNIKExists
	}

	// Check if email already exists (if provided)
	if input.Email != "" {
		emailExists, err := s.store.CheckEmailExists(ctx, pgtype.Text{String: input.Email, Valid: true})
		if err != nil {
			return nil, ErrInternalError
		}
		if emailExists {
			return nil, ErrEmailExists
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternalError
	}

	// Build params
	params := pg_store.CreatePendudukParams{
		Nik:          input.NIK,
		NamaLengkap:  input.NamaLengkap,
		JenisKelamin: input.JenisKelamin,
		KelurahanID:  pgtype.Int2{Int16: input.KelurahanID, Valid: true},
		PasswordHash: pgtype.Text{String: string(hashedPassword), Valid: true},
	}

	if input.Email != "" {
		params.Email = pgtype.Text{String: input.Email, Valid: true}
	}
	if input.Alamat != "" {
		params.Alamat = pgtype.Text{String: input.Alamat, Valid: true}
	}
	if input.NoHP != "" {
		params.NoHp = pgtype.Text{String: input.NoHP, Valid: true}
	}

	// Create penduduk
	penduduk, err := s.store.CreatePenduduk(ctx, params)
	if err != nil {
		return nil, ErrInternalError
	}

	return &RegisterResult{
		NIK:         penduduk.Nik,
		NamaLengkap: penduduk.NamaLengkap,
	}, nil
}
