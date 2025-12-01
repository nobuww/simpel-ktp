-- name: GetPendudukByNIK :one
SELECT * FROM penduduk WHERE nik = $1;

-- name: CreatePenduduk :one
INSERT INTO penduduk (
    nik,
    kelurahan_id,
    email,
    password_hash,
    nama_lengkap,
    alamat,
    no_hp,
    jenis_kelamin
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPetugasByNIP :one
SELECT * FROM petugas WHERE nip = $1 AND is_active = TRUE;

-- name: CheckPendudukExists :one
SELECT EXISTS(SELECT 1 FROM penduduk WHERE nik = $1) AS exists;

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM penduduk WHERE email = $1) AS exists;
