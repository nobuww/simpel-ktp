package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/nobuww/simpel-ktp/internal/session"
	"github.com/nobuww/simpel-ktp/internal/store"
)

type Handler struct {
	store   *store.Store
	service *Service
	session *session.Manager
}

func New(store *store.Store, sessionMgr *session.Manager) *Handler {
	return &Handler{
		store:   store,
		service: NewService(store),
		session: sessionMgr,
	}
}

func (h *Handler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	LoginPage().Render(r.Context(), w)
}

func (h *Handler) LoginPetugasPageHandler(w http.ResponseWriter, r *http.Request) {
	LoginPetugasPage().Render(r.Context(), w)
}

func (h *Handler) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch kelurahan list from database
	kelurahanList, err := h.store.ListAllKelurahan(ctx)
	if err != nil {
		kelurahanList = nil
	}

	// Convert to template-friendly format
	kelurahanOptions := make([]KelurahanOption, 0, len(kelurahanList))
	for _, k := range kelurahanList {
		kelurahanOptions = append(kelurahanOptions, KelurahanOption{
			ID:   k.ID,
			Nama: k.NamaKelurahan,
		})
	}

	RegisterPage(kelurahanOptions).Render(ctx, w)
}

// HandleLogin handles POST /auth/login for warga (citizen) login
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		AuthError("Gagal memproses form").Render(ctx, w)
		return
	}

	nik := strings.TrimSpace(r.FormValue("nik"))
	password := r.FormValue("password")
	remember := r.FormValue("remember") == "on"

	// Validate input
	if nik == "" || password == "" {
		AuthError("NIK dan password harus diisi").Render(ctx, w)
		return
	}
	if len(nik) != 16 {
		AuthError("NIK harus 16 digit").Render(ctx, w)
		return
	}

	// Call service
	result, err := h.service.LoginWarga(ctx, WargaLoginInput{
		NIK:      nik,
		Password: password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			AuthError("NIK atau password salah").Render(ctx, w)
		case errors.Is(err, ErrAccountNoPassword):
			AuthError("Akun belum memiliki password. Silakan hubungi petugas.").Render(ctx, w)
		default:
			AuthError("Terjadi kesalahan, silakan coba lagi").Render(ctx, w)
		}
		return
	}

	// Create session
	if err := h.session.SetWargaSession(w, r, result.NIK, result.NamaLengkap, remember); err != nil {
		AuthError("Gagal membuat sesi login").Render(ctx, w)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	AuthSuccess("Login berhasil! Mengalihkan...").Render(ctx, w)
}

// HandleLoginPetugas handles POST /auth/login/petugas for petugas (officer) login
func (h *Handler) HandleLoginPetugas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		AuthError("Gagal memproses form").Render(ctx, w)
		return
	}

	nip := strings.TrimSpace(r.FormValue("nip"))
	password := r.FormValue("password")
	remember := r.FormValue("remember") == "on"

	// Validate input
	if nip == "" || password == "" {
		AuthError("NIP dan password harus diisi").Render(ctx, w)
		return
	}
	if len(nip) != 18 {
		AuthError("NIP harus 18 digit").Render(ctx, w)
		return
	}

	// Call service
	result, err := h.service.LoginPetugas(ctx, PetugasLoginInput{
		NIP:      nip,
		Password: password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			AuthError("NIP atau password salah").Render(ctx, w)
		default:
			AuthError("Terjadi kesalahan, silakan coba lagi").Render(ctx, w)
		}
		return
	}

	// Create session
	if err := h.session.SetPetugasSession(w, r, result.ID, result.NamaPetugas, result.Role, remember); err != nil {
		AuthError("Gagal membuat sesi login").Render(ctx, w)
		return
	}

	w.Header().Set("HX-Redirect", "/admin")
	AuthSuccess("Login berhasil! Mengalihkan ke dashboard...").Render(ctx, w)
}

// HandleRegister handles POST /auth/register for warga (citizen) registration
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		AuthError("Gagal memproses form").Render(ctx, w)
		return
	}

	// Extract and validate form values
	nik := strings.TrimSpace(r.FormValue("nik"))
	namaLengkap := strings.TrimSpace(r.FormValue("nama_lengkap"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	alamat := strings.TrimSpace(r.FormValue("alamat"))
	noHP := strings.TrimSpace(r.FormValue("no_hp"))
	jenisKelamin := r.FormValue("jenis_kelamin")
	kelurahanIDStr := r.FormValue("kelurahan_id")

	// Validate required fields
	if nik == "" {
		AuthError("NIK harus diisi").Render(ctx, w)
		return
	}
	if len(nik) != 16 {
		AuthError("NIK harus 16 digit").Render(ctx, w)
		return
	}
	if namaLengkap == "" {
		AuthError("Nama lengkap harus diisi").Render(ctx, w)
		return
	}
	if password == "" {
		AuthError("Password harus diisi").Render(ctx, w)
		return
	}
	if len(password) < 8 {
		AuthError("Password minimal 8 karakter").Render(ctx, w)
		return
	}
	if jenisKelamin != "LAKI_LAKI" && jenisKelamin != "PEREMPUAN" {
		AuthError("Jenis kelamin tidak valid").Render(ctx, w)
		return
	}
	if kelurahanIDStr == "" {
		AuthError("Kelurahan harus dipilih").Render(ctx, w)
		return
	}

	kelurahanID, err := strconv.ParseInt(kelurahanIDStr, 10, 16)
	if err != nil {
		AuthError("ID Kelurahan tidak valid").Render(ctx, w)
		return
	}

	// Call service
	result, err := h.service.RegisterWarga(ctx, RegisterInput{
		NIK:          nik,
		NamaLengkap:  namaLengkap,
		Email:        email,
		Password:     password,
		Alamat:       alamat,
		NoHP:         noHP,
		JenisKelamin: jenisKelamin,
		KelurahanID:  int16(kelurahanID),
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrNIKExists):
			AuthError("NIK sudah terdaftar").Render(ctx, w)
		case errors.Is(err, ErrEmailExists):
			AuthError("Email sudah terdaftar").Render(ctx, w)
		default:
			AuthError("Gagal mendaftarkan akun. Silakan coba lagi.").Render(ctx, w)
		}
		return
	}

	// Auto-login after registration
	if err := h.session.SetWargaSession(w, r, result.NIK, result.NamaLengkap, false); err != nil {
		w.Header().Set("HX-Redirect", "/login")
		AuthSuccess("Pendaftaran berhasil! Silakan login.").Render(ctx, w)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	AuthSuccess("Pendaftaran berhasil! Mengalihkan...").Render(ctx, w)
}

// HandleLogout handles POST /auth/logout
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.session.ClearSession(w, r); err != nil {
		http.Error(w, "Gagal logout", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
