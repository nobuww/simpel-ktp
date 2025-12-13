package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/internal/store/pg_store"
)

type kelurahanSeed struct {
	Nama     string
	KodeArea string
}

type petugasSeed struct {
	NIP         string
	Nama        string
	Username    string
	Password    string
	KelurahanID *int16
}

type pendudukSeed struct {
	NIK          string
	Nama         string
	Email        string
	Alamat       string
	NoHP         string
	JenisKelamin string
	KodeArea     string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// CLI flags: --reset to truncate seed tables, --force to allow running in production
	reset := flag.Bool("reset", false, "Truncate seed tables before seeding (destructive)")
	force := flag.Bool("force", false, "Force destructive operations even in production")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to db: %v\n", err)
	}
	defer dbPool.Close()

	s := store.New(dbPool)

	fmt.Println("Starting database seeding...")

	if *reset {
		goEnv := os.Getenv("GO_ENV")
		if goEnv == "production" && !*force {
			log.Fatalf("Refusing to reset database in production. Use --force to override.")
		}

		fmt.Println("\nResetting seed tables (this will remove existing data)...")
		err := s.ExecTx(ctx, func(q *pg_store.Queries) error {
			return q.TruncateSeedTables(ctx)
		})
		if err != nil {
			log.Fatalf("Failed to reset db: %v", err)
		}
		fmt.Println("Reset completed — database cleaned for seeding.")
	}

	kelurahanList := []kelurahanSeed{
		{Nama: "Pademangan Barat", KodeArea: "PMB"},
		{Nama: "Pademangan Timur", KodeArea: "PMT"},
		{Nama: "Ancol", KodeArea: "ACL"},
	}

	kelurahanIDs := make(map[string]int16)

	fmt.Println("\nSeeding Kelurahan...")
	for _, k := range kelurahanList {
		existing, err := s.GetKelurahanByKodeArea(ctx, k.KodeArea)
		if err == nil {
			fmt.Printf("   ✓ Kelurahan %s already exists (ID: %d)\n", k.Nama, existing.ID)
			kelurahanIDs[k.KodeArea] = existing.ID
			continue
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			log.Fatalf("Failed to query kelurahan %s: %v", k.KodeArea, err)
		}

		created, err := s.CreateKelurahan(ctx, pg_store.CreateKelurahanParams{
			NamaKelurahan: k.Nama,
			KodeArea:      k.KodeArea,
		})
		if err != nil {
			log.Fatalf("Failed to create kelurahan %s: %v", k.Nama, err)
		}
		fmt.Printf("   ✓ Created kelurahan: %s (ID: %d, Kode: %s)\n", created.NamaKelurahan, created.ID, created.KodeArea)
		kelurahanIDs[k.KodeArea] = created.ID
	}

	petugasList := []petugasSeed{
		{
			NIP:         "198501152010011001",
			Nama:        "Budi Santoso",
			Username:    "admin.kecamatan",
			Password:    "admin123",
			KelurahanID: nil,
		},
		{
			NIP:         "199003202015012001",
			Nama:        "Siti Rahayu",
			Username:    "admin.pademanganbarat",
			Password:    "admin123",
			KelurahanID: ptr(kelurahanIDs["PMB"]),
		},
		{
			NIP:         "198807112012011002",
			Nama:        "Ahmad Hidayat",
			Username:    "admin.pademangantimur",
			Password:    "admin123",
			KelurahanID: ptr(kelurahanIDs["PMT"]),
		},
		{
			NIP:         "199205182018012003",
			Nama:        "Dewi Lestari",
			Username:    "admin.ancol",
			Password:    "admin123",
			KelurahanID: ptr(kelurahanIDs["ACL"]),
		},
	}

	fmt.Println("\n👤 Seeding Petugas...")
	for _, p := range petugasList {
		_, err := s.GetPetugasByUsername(ctx, p.Username)
		if err == nil {
			fmt.Printf("   ✓ Petugas %s already exists\n", p.Username)
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Fatalf("Failed to query petugas %s: %v", p.Username, err)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		params := pg_store.CreatePetugasParams{
			NamaPetugas:  p.Nama,
			Username:     p.Username,
			PasswordHash: string(hashedPassword),
		}

		params.Nip = pgtype.Text{String: p.NIP, Valid: true}

		if p.KelurahanID != nil {
			if *p.KelurahanID == 0 {
				log.Fatalf("Invalid kelurahan ID for petugas %s: id is 0 — check kelurahan mapping", p.Username)
			}
			params.KelurahanID = pgtype.Int2{Int16: *p.KelurahanID, Valid: true}
		}

		created, err := s.CreatePetugas(ctx, params)
		if err != nil {
			log.Fatalf("Failed to create petugas %s: %v", p.Username, err)
		}

		role := "ADMIN_KECAMATAN"
		if p.KelurahanID != nil {
			role = "ADMIN_KELURAHAN"
		}
		fmt.Printf("   ✓ Created petugas: %s (%s) - Role: %s\n", created.NamaPetugas, created.Username, role)
	}

	fmt.Println("\nDatabase seeding completed!")
	fmt.Println("\nLogin Credentials:")
	fmt.Println("   ─────────────────────────────────────────")
	for _, p := range petugasList {
		role := "ADMIN_KECAMATAN"
		if p.KelurahanID != nil {
			role = "ADMIN_KELURAHAN"
		}
		fmt.Printf("   Username: %-25s Password: %s (%s)\n", p.Username, p.Password, role)
	}

	// Data Penduduk (15 entries)
	pendudukList := []pendudukSeed{
		{NIK: "3172010101800001", Nama: "Ahmad Supriyadi", Email: "ahmad.supriyadi@example.com", Alamat: "Jl. Pademangan II Gg. 2 No. 10", NoHP: "081234567890", JenisKelamin: "LAKI_LAKI", KodeArea: "PMB"},
		{NIK: "3172010101850002", Nama: "Siti Aminah", Email: "siti.aminah@example.com", Alamat: "Jl. Pademangan III No. 5", NoHP: "081234567891", JenisKelamin: "PEREMPUAN", KodeArea: "PMB"},
		{NIK: "3172010101900003", Nama: "Budi Harsono", Email: "budi.harsono@example.com", Alamat: "Jl. Pademangan IV No. 12", NoHP: "081234567892", JenisKelamin: "LAKI_LAKI", KodeArea: "PMB"},
		{NIK: "3172010101950004", Nama: "Dewi Kartika", Email: "dewi.kartika@example.com", Alamat: "Jl. Budi Mulia No. 8", NoHP: "081234567893", JenisKelamin: "PEREMPUAN", KodeArea: "PMB"},
		{NIK: "3172010101880005", Nama: "Eko Prasetyo", Email: "eko.prasetyo@example.com", Alamat: "Jl. Hidup Baru No. 3", NoHP: "081234567894", JenisKelamin: "LAKI_LAKI", KodeArea: "PMB"},
		{NIK: "3172020202820001", Nama: "Fajar Nugraha", Email: "fajar.nugraha@example.com", Alamat: "Jl. Pademangan Timur VIII No. 20", NoHP: "081234567895", JenisKelamin: "LAKI_LAKI", KodeArea: "PMT"},
		{NIK: "3172020202870002", Nama: "Gita Permata", Email: "gita.permata@example.com", Alamat: "Jl. Pademangan Timur IX No. 15", NoHP: "081234567896", JenisKelamin: "PEREMPUAN", KodeArea: "PMT"},
		{NIK: "3172020202920003", Nama: "Hendra Wijaya", Email: "hendra.wijaya@example.com", Alamat: "Jl. Pademangan Timur X No. 7", NoHP: "081234567897", JenisKelamin: "LAKI_LAKI", KodeArea: "PMT"},
		{NIK: "3172020202970004", Nama: "Indah Sari", Email: "indah.sari@example.com", Alamat: "Jl. Pademangan Timur XI No. 4", NoHP: "081234567898", JenisKelamin: "PEREMPUAN", KodeArea: "PMT"},
		{NIK: "3172020202850005", Nama: "Joko Susilo", Email: "joko.susilo@example.com", Alamat: "Jl. Pademangan Timur XII No. 9", NoHP: "081234567899", JenisKelamin: "LAKI_LAKI", KodeArea: "PMT"},
		{NIK: "3172030303830001", Nama: "Kartini Putri", Email: "kartini.putri@example.com", Alamat: "Jl. Lodan Raya No. 10", NoHP: "081234567900", JenisKelamin: "PEREMPUAN", KodeArea: "ACL"},
		{NIK: "3172030303880002", Nama: "Lukman Hakim", Email: "lukman.hakim@example.com", Alamat: "Jl. Pasir Putih No. 5", NoHP: "081234567901", JenisKelamin: "LAKI_LAKI", KodeArea: "ACL"},
		{NIK: "3172030303930003", Nama: "Maya Puspita", Email: "maya.puspita@example.com", Alamat: "Jl. Pantai Indah No. 12", NoHP: "081234567902", JenisKelamin: "PEREMPUAN", KodeArea: "ACL"},
		{NIK: "3172030303980004", Nama: "Nurhadi Surya", Email: "nurhadi.surya@example.com", Alamat: "Jl. Ancol Barat No. 8", NoHP: "081234567903", JenisKelamin: "LAKI_LAKI", KodeArea: "ACL"},
		{NIK: "3172030303860005", Nama: "Oki Pradana", Email: "oki.pradana@example.com", Alamat: "Jl. Ancol Timur No. 3", NoHP: "081234567904", JenisKelamin: "LAKI_LAKI", KodeArea: "ACL"},
	}

	fmt.Println("\n👥 Seeding Penduduk...")
	pendudukPassword := "rahasia123"
	hashedPendudukPassword, err := bcrypt.GenerateFromPassword([]byte(pendudukPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash penduduk password: %v", err)
	}

	for _, p := range pendudukList {
		exists, err := s.CheckPendudukExists(ctx, p.NIK)
		if err != nil {
			log.Fatalf("Failed to check penduduk existence %s: %v", p.NIK, err)
		}
		if exists {
			fmt.Printf("   ✓ Penduduk %s already exists\n", p.Nama)
			continue
		}

		kelID, ok := kelurahanIDs[p.KodeArea]
		if !ok {
			log.Fatalf("Kode area %s not found for penduduk %s", p.KodeArea, p.Nama)
		}

		created, err := s.CreatePenduduk(ctx, pg_store.CreatePendudukParams{
			Nik:          p.NIK,
			KelurahanID:  pgtype.Int2{Int16: kelID, Valid: true},
			Email:        pgtype.Text{String: p.Email, Valid: true},
			PasswordHash: pgtype.Text{String: string(hashedPendudukPassword), Valid: true},
			NamaLengkap:  p.Nama,
			Alamat:       pgtype.Text{String: p.Alamat, Valid: true},
			NoHp:         pgtype.Text{String: p.NoHP, Valid: true},
			JenisKelamin: p.JenisKelamin,
		})
		if err != nil {
			log.Fatalf("Failed to create penduduk %s: %v", p.Nama, err)
		}

		fmt.Printf("   ✓ Created penduduk: %s (%s) - %s\n", created.NamaLengkap, created.Nik, p.KodeArea)
	}

	fmt.Println("\nPenduduk Credentials (All):")
	fmt.Println("   Password: rahasia123")
	fmt.Println("   ─────────────────────────────────────────")
}

func ptr(v int16) *int16 {
	return &v
}
