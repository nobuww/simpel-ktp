-- name: ListJadwalSesi :many
SELECT 
    js.id,
    js.tanggal,
    js.jam_mulai,
    js.jam_selesai,
    js.kuota_terisi,
    js.kuota_maksimal,
    js.status_sesi,
    k.nama_kelurahan
FROM jadwal_sesi js
JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
WHERE js.tanggal >= $1 AND js.tanggal <= $2
  AND ($3::integer IS NULL OR js.lokasi_kelurahan_id = $3)
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
    k.nama_kelurahan
FROM jadwal_sesi js
JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
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
