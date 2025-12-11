-- name: ListJadwalSesi :many
SELECT 
    js.id,
    js.tanggal,
    js.jam_mulai,
    js.jam_selesai,
    js.kuota_terisi,
    js.kuota_maksimal,
    js.status_sesi,
    COALESCE(k.nama_kelurahan, 'Kecamatan Pademangan')::text as nama_kelurahan
FROM jadwal_sesi js
LEFT JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
WHERE js.tanggal >= $1 AND js.tanggal <= $2
  AND (
    (sqlc.narg('kelurahan_id')::integer IS NOT NULL AND js.lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
    OR
    (sqlc.narg('kelurahan_id')::integer IS NULL AND js.lokasi_kelurahan_id IS NULL)
  )
ORDER BY js.tanggal, js.jam_mulai;

-- name: GetJadwalSesiById :one
SELECT 
    js.id,
    js.tanggal,
    js.jam_mulai,
    js.jam_selesai,
    js.kuota_terisi,
    js.kuota_maksimal,
    js.status_sesi,
    js.lokasi_kelurahan_id,
    COALESCE(k.nama_kelurahan, 'Kecamatan Pademangan')::text as nama_kelurahan
FROM jadwal_sesi js
LEFT JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
WHERE js.id = $1;

-- name: CreateJadwalSesi :one
INSERT INTO jadwal_sesi (
    lokasi_kelurahan_id,
    tanggal,
    jam_mulai,
    jam_selesai,
    kuota_maksimal,
    kuota_terisi,
    status_sesi
) VALUES ($1, $2, $3, $4, $5, 0, 'BUKA')
ON CONFLICT (tanggal, jam_mulai, lokasi_kelurahan_id) DO NOTHING
RETURNING id;

-- name: UpdateJadwalSesi :exec
UPDATE jadwal_sesi
SET 
    tanggal = $2,
    jam_mulai = $3,
    jam_selesai = $4,
    kuota_maksimal = $5,
    status_sesi = $6
WHERE id = $1;

-- name: IncrementKuotaTerisi :exec
UPDATE jadwal_sesi
SET kuota_terisi = kuota_terisi + 1
WHERE id = $1 AND kuota_terisi < kuota_maksimal;

-- name: ListAllKelurahan :many
SELECT id, nama_kelurahan
FROM ref_kelurahan
ORDER BY nama_kelurahan;

-- name: GetKelurahanById :one
SELECT id, nama_kelurahan
FROM ref_kelurahan
WHERE id = $1;

-- name: ListPermohonanByJadwal :many
SELECT 
    p.id,
    p.kode_booking,
    p.nik,
    pd.nama_lengkap,
    p.status_terkini,
    p.nomor_antrian_sesi as nomor_antrian
FROM permohonan p
JOIN penduduk pd ON p.nik = pd.nik
WHERE p.jadwal_sesi_id = $1
ORDER BY p.nomor_antrian_sesi ASC;

-- name: DeleteJadwalSesi :exec
DELETE FROM jadwal_sesi
WHERE id = $1
  AND (
    (sqlc.narg('kelurahan_id')::integer IS NOT NULL AND lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
    OR
    (sqlc.narg('kelurahan_id')::integer IS NULL AND lokasi_kelurahan_id IS NULL)
  );

-- name: CountPermohonanByJadwal :one
SELECT COUNT(*) as count
FROM permohonan
WHERE jadwal_sesi_id = $1;
