-- name: CreateKelurahan :one
INSERT INTO ref_kelurahan (nama_kelurahan, kode_area)
VALUES ($1, $2)
RETURNING *;

-- name: GetKelurahanByKodeArea :one
SELECT * FROM ref_kelurahan WHERE kode_area = $1;

-- name: CreatePetugas :one
INSERT INTO petugas (kelurahan_id, nip, nama_petugas, created_by, username, password_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPetugasByUsername :one
SELECT * FROM petugas WHERE username = $1;

-- name: ListKelurahan :many
SELECT * FROM ref_kelurahan ORDER BY id;

-- name: TruncateSeedTables :exec
TRUNCATE riwayat_status, dokumen_syarat, permohonan, jadwal_sesi, petugas, penduduk, ref_kelurahan RESTART IDENTITY CASCADE;
