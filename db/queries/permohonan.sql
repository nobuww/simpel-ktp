-- name: GetPermohonanStatusById :one
SELECT 
    p.id,
    p.status_terkini
FROM permohonan p
WHERE p.id = $1;

-- name: UpdatePermohonanStatus :exec
UPDATE permohonan
SET status_terkini = $2, updated_at = NOW()
WHERE id = $1;

-- name: InsertRiwayatStatus :exec
INSERT INTO riwayat_status (
    permohonan_id,
    petugas_id,
    status_baru,
    catatan_proses,
    waktu_proses
) VALUES ($1, $2, $3, $4, NOW());

-- name: ListPermohonanByStatus :many
SELECT 
    p.id,
    p.kode_booking,
    pd.nik,
    pd.nama_lengkap,
    p.jenis_permohonan,
    p.status_terkini,
    p.created_at as tanggal_daftar,
    js.tanggal as jadwal_tanggal,
    js.jam_mulai as jadwal_jam,
    p.nomor_antrian_sesi as nomor_antrian
FROM permohonan p
JOIN penduduk pd ON p.nik = pd.nik
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
WHERE ($1::text IS NULL OR p.status_terkini = $1)
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPermohonanByStatus :one
SELECT 
    COUNT(*) FILTER (WHERE status_terkini = 'VERIFIKASI') as verifikasi,
    COUNT(*) FILTER (WHERE status_terkini = 'PROSES') as proses,
    COUNT(*) FILTER (WHERE status_terkini = 'SIAP_AMBIL') as siap_ambil,
    COUNT(*) FILTER (WHERE status_terkini = 'SELESAI') as selesai,
    COUNT(*) FILTER (WHERE status_terkini = 'DITOLAK') as ditolak,
    COUNT(*) as total
FROM permohonan;

-- name: GetPermohonanDetail :one
SELECT 
    p.id,
    p.kode_booking,
    p.jenis_permohonan,
    p.status_terkini,
    p.nomor_antrian_sesi as nomor_antrian,
    p.created_at,
    pd.nik,
    pd.nama_lengkap,
    pd.jenis_kelamin,
    pd.alamat,
    pd.no_hp,
    COALESCE(k.nama_kelurahan, 'Kecamatan Pademangan')::text as nama_kelurahan,
    js.tanggal as jadwal_tanggal,
    js.jam_mulai as jadwal_jam_mulai,
    js.jam_selesai as jadwal_jam_selesai
FROM permohonan p
JOIN penduduk pd ON p.nik = pd.nik
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
LEFT JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
WHERE p.id = $1;

-- name: GetRiwayatStatusByPermohonan :many
SELECT 
    rs.status_baru,
    rs.catatan_proses,
    rs.waktu_proses,
    pt.nama_petugas
FROM riwayat_status rs
LEFT JOIN petugas pt ON rs.petugas_id = pt.id
WHERE rs.permohonan_id = $1
ORDER BY rs.waktu_proses DESC;

-- name: CreatePermohonan :one
INSERT INTO permohonan (
    nik,
    jadwal_sesi_id,
    jenis_permohonan
) VALUES ($1, $2, $3)
RETURNING id;

-- name: CreateDokumenSyarat :exec
INSERT INTO dokumen_syarat (
    permohonan_id,
    file_path,
    jenis_dokumen
) VALUES ($1, $2, $3);
