-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Table: ref_kelurahan
CREATE TABLE ref_kelurahan (
    id SMALLSERIAL PRIMARY KEY,
    nama_kelurahan TEXT NOT NULL,
    kode_area CHAR(3) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: penduduk
CREATE TABLE penduduk (
    nik CHAR(16) PRIMARY KEY,
    kelurahan_id SMALLINT REFERENCES ref_kelurahan(id),
    email TEXT UNIQUE,
    password_hash TEXT,
    nama_lengkap TEXT NOT NULL,
    alamat TEXT,
    no_hp TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    jenis_kelamin TEXT NOT NULL,
    CONSTRAINT chk_jenis_kelamin CHECK (jenis_kelamin IN ('LAKI_LAKI', 'PEREMPUAN'))
);

-- Table: petugas
CREATE TABLE petugas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kelurahan_id SMALLINT REFERENCES ref_kelurahan(id),
    nip CHAR(18) UNIQUE,
    nama_petugas TEXT NOT NULL,
    created_by UUID REFERENCES petugas(id),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    role TEXT NOT NULL,
    CONSTRAINT chk_petugas_role CHECK (role IN ('ADMIN_KECAMATAN', 'ADMIN_KELURAHAN'))
);

-- Table: jadwal_sesi
CREATE TABLE jadwal_sesi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lokasi_kelurahan_id SMALLINT REFERENCES ref_kelurahan(id),
    tanggal DATE NOT NULL,
    jam_mulai TIME NOT NULL,
    jam_selesai TIME NOT NULL,
    kuota_maksimal SMALLINT NOT NULL DEFAULT 50,
    kuota_terisi SMALLINT NOT NULL DEFAULT 0,
    
    status_sesi TEXT DEFAULT 'BUKA',
    CONSTRAINT chk_status_sesi CHECK (status_sesi IN ('BUKA', 'PENUH', 'ISTIRAHAT', 'TUTUP')),

    CONSTRAINT unique_sesi_lokasi UNIQUE (tanggal, jam_mulai, lokasi_kelurahan_id)
);

-- Table: permohonan
CREATE TABLE permohonan (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nik CHAR(16) REFERENCES penduduk(nik),
    jadwal_sesi_id UUID REFERENCES jadwal_sesi(id),
    kode_booking CHAR(10) UNIQUE, 
    nomor_antrian_sesi SMALLINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    jenis_permohonan TEXT NOT NULL,
    CONSTRAINT chk_jenis_permohonan CHECK (jenis_permohonan IN ('BARU', 'HILANG', 'RUSAK', 'UPDATE')),

    status_terkini TEXT DEFAULT 'VERIFIKASI',
    CONSTRAINT chk_status_permohonan CHECK (status_terkini IN ('VERIFIKASI', 'PROSES', 'SIAP_AMBIL', 'SELESAI', 'DITOLAK')),

    CONSTRAINT unique_antrian_per_sesi UNIQUE (jadwal_sesi_id, nomor_antrian_sesi)
);

-- Table: dokumen_syarat
CREATE TABLE dokumen_syarat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permohonan_id UUID REFERENCES permohonan(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    jenis_dokumen TEXT NOT NULL,
    CONSTRAINT chk_jenis_dokumen CHECK (jenis_dokumen IN ('KK', 'SURAT_POLISI', 'KTP_RUSAK', 'PENGANTAR'))
);

-- Table: riwayat_status
CREATE TABLE riwayat_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permohonan_id UUID REFERENCES permohonan(id) ON DELETE CASCADE,
    petugas_id UUID REFERENCES petugas(id),
    status_baru TEXT NOT NULL,
    catatan_proses TEXT,
    waktu_proses TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create Indexes
CREATE INDEX idx_penduduk_search ON penduduk(nama_lengkap, email);
CREATE INDEX idx_permohonan_status ON permohonan(status_terkini);
CREATE INDEX idx_permohonan_sesi ON permohonan(jadwal_sesi_id);
CREATE INDEX idx_dokumen_permohonan ON dokumen_syarat(permohonan_id);

-- Function: Auto-Assign Role
CREATE OR REPLACE FUNCTION set_petugas_role()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.kelurahan_id IS NULL THEN
        NEW.role := 'ADMIN_KECAMATAN';
    ELSE
        NEW.role := 'ADMIN_KELURAHAN';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function: Booking Logic (Quota check & Code Generation)
CREATE OR REPLACE FUNCTION process_new_permohonan()
RETURNS TRIGGER AS $$
DECLARE
    v_kode_area CHAR(3);
    v_tanggal_day TEXT;
    v_random_str TEXT;
    v_kuota_max INT;
    v_kuota_now INT;
BEGIN
    -- Lock row for concurrency safety
    SELECT 
        k.kode_area, 
        TO_CHAR(j.tanggal, 'DD'),
        j.kuota_maksimal,
        j.kuota_terisi
    INTO v_kode_area, v_tanggal_day, v_kuota_max, v_kuota_now
    FROM jadwal_sesi j
    JOIN ref_kelurahan k ON j.lokasi_kelurahan_id = k.id
    WHERE j.id = NEW.jadwal_sesi_id
    FOR UPDATE;

    -- Validation
    IF v_kuota_now >= v_kuota_max THEN
        RAISE EXCEPTION 'Session is full (Quota Reached)';
    END IF;

    -- Update Session
    UPDATE jadwal_sesi 
    SET kuota_terisi = kuota_terisi + 1,
        status_sesi = CASE WHEN (kuota_terisi + 1) >= kuota_maksimal THEN 'PENUH' ELSE status_sesi END
    WHERE id = NEW.jadwal_sesi_id;

    -- Set Queue Number
    NEW.nomor_antrian_sesi := v_kuota_now + 1;

    -- Generate Booking Code
    v_random_str := substring(md5(random()::text), 1, 4);
    NEW.kode_booking := UPPER(v_kode_area || '-' || v_tanggal_day || '-' || v_random_str);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function: Auto-Log Status History
CREATE OR REPLACE FUNCTION log_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') OR (OLD.status_terkini IS DISTINCT FROM NEW.status_terkini) THEN
        INSERT INTO riwayat_status (
            permohonan_id, 
            status_baru, 
            catatan_proses, 
            waktu_proses
        ) VALUES (
            NEW.id, 
            NEW.status_terkini, 
            CASE WHEN TG_OP = 'INSERT' THEN 'Permohonan dibuat' ELSE 'Status updated automatically' END,
            NOW()
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach Triggers
CREATE TRIGGER trg_set_petugas_role
BEFORE INSERT OR UPDATE OF kelurahan_id ON petugas
FOR EACH ROW EXECUTE FUNCTION set_petugas_role();

CREATE TRIGGER trg_process_new_permohonan
BEFORE INSERT ON permohonan
FOR EACH ROW EXECUTE FUNCTION process_new_permohonan();

CREATE TRIGGER trg_log_status_change
AFTER INSERT OR UPDATE ON permohonan
FOR EACH ROW EXECUTE FUNCTION log_status_change();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop Triggers
DROP TRIGGER IF EXISTS trg_log_status_change ON permohonan;
DROP TRIGGER IF EXISTS trg_process_new_permohonan ON permohonan;
DROP TRIGGER IF EXISTS trg_set_petugas_role ON petugas;

-- Drop Functions
DROP FUNCTION IF EXISTS log_status_change;
DROP FUNCTION IF EXISTS process_new_permohonan;
DROP FUNCTION IF EXISTS set_petugas_role;

-- Drop Tables (Reverse Order of Dependencies)
DROP TABLE IF EXISTS riwayat_status;
DROP TABLE IF EXISTS dokumen_syarat;
DROP TABLE IF EXISTS permohonan;
DROP TABLE IF EXISTS jadwal_sesi;
DROP TABLE IF EXISTS petugas;
DROP TABLE IF EXISTS penduduk;
DROP TABLE IF EXISTS ref_kelurahan;

-- Drop Extension
DROP EXTENSION IF EXISTS "pgcrypto";

-- +goose StatementEnd