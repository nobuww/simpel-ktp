-- name: GetAdminDashboardStats :one
SELECT 
    COUNT(*) AS total_permohonan,
    COUNT(*) FILTER (WHERE status_terkini = 'VERIFIKASI') AS verifikasi,
    COUNT(*) FILTER (WHERE status_terkini = 'PROSES') AS proses,
    COUNT(*) FILTER (WHERE status_terkini = 'SIAP_AMBIL') AS siap_ambil,
    COUNT(*) FILTER (WHERE status_terkini = 'SELESAI') AS selesai,
    COUNT(*) FILTER (WHERE status_terkini = 'DITOLAK') AS ditolak
FROM permohonan p
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
WHERE (
    -- Admin Kelurahan: filter by their kelurahan
    (sqlc.narg('kelurahan_id')::smallint IS NOT NULL AND js.lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
    OR
    -- Admin Kecamatan: only see kecamatan-level data (NULL lokasi_kelurahan_id)
    (sqlc.narg('kelurahan_id')::smallint IS NULL AND js.lokasi_kelurahan_id IS NULL)
);

-- name: ListPermohonanAdmin :many
SELECT 
    p.id,
    p.kode_booking,
    p.nik,
    pd.nama_lengkap,
    p.jenis_permohonan,
    p.status_terkini,
    p.created_at as tanggal_daftar,
    js.tanggal as jadwal_tanggal,
    js.jam_mulai as jadwal_jam_mulai,
    js.jam_selesai as jadwal_jam_selesai,
    p.nomor_antrian_sesi as nomor_antrian,
    k.id as lokasi_kelurahan_id,
    k.nama_kelurahan as lokasi_nama_kelurahan
FROM permohonan p
JOIN penduduk pd ON p.nik = pd.nik
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
LEFT JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
WHERE 
    (sqlc.narg('search')::text IS NULL OR 
     p.nik ILIKE '%' || sqlc.narg('search') || '%' OR 
     p.kode_booking ILIKE '%' || sqlc.narg('search') || '%' OR 
     pd.nama_lengkap ILIKE '%' || sqlc.narg('search') || '%')
    AND (sqlc.narg('status')::text IS NULL OR p.status_terkini = sqlc.narg('status'))
    AND (
        (sqlc.narg('kelurahan_id')::smallint IS NOT NULL AND js.lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
        OR
        (sqlc.narg('kelurahan_id')::smallint IS NULL AND js.lokasi_kelurahan_id IS NULL)
    )
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPermohonanAdmin :one
SELECT COUNT(*) 
FROM permohonan p
JOIN penduduk pd ON p.nik = pd.nik
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
WHERE 
    (sqlc.narg('search')::text IS NULL OR 
     p.nik ILIKE '%' || sqlc.narg('search') || '%' OR 
     p.kode_booking ILIKE '%' || sqlc.narg('search') || '%' OR 
     pd.nama_lengkap ILIKE '%' || sqlc.narg('search') || '%')
    AND (sqlc.narg('status')::text IS NULL OR p.status_terkini = sqlc.narg('status'))
    AND (
        (sqlc.narg('kelurahan_id')::smallint IS NOT NULL AND js.lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
        OR
        (sqlc.narg('kelurahan_id')::smallint IS NULL AND js.lokasi_kelurahan_id IS NULL)
    );

-- name: GetPermohonanDetailAdmin :one
SELECT 
    p.id,
    p.kode_booking,
    p.nik,
    pd.nama_lengkap,
    pd.jenis_kelamin,
    pd.alamat,
    pd.no_hp as no_telp,
    k_penduduk.nama_kelurahan as kelurahan,
    p.jenis_permohonan,
    p.status_terkini,
    p.created_at as tanggal_daftar,
    js.tanggal as jadwal_tanggal,
    js.jam_mulai as jadwal_jam_mulai,
    js.jam_selesai as jadwal_jam_selesai,
    p.nomor_antrian_sesi as nomor_antrian,
    k_lokasi.nama_kelurahan as lokasi_permohonan
FROM permohonan p
JOIN penduduk pd ON p.nik = pd.nik
LEFT JOIN ref_kelurahan k_penduduk ON pd.kelurahan_id = k_penduduk.id
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
LEFT JOIN ref_kelurahan k_lokasi ON js.lokasi_kelurahan_id = k_lokasi.id
WHERE p.id = $1
  AND (
    (sqlc.narg('kelurahan_id')::smallint IS NOT NULL AND js.lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
    OR
    (sqlc.narg('kelurahan_id')::smallint IS NULL AND js.lokasi_kelurahan_id IS NULL)
  );

-- name: UpdatePermohonanStatusAdmin :exec
UPDATE permohonan p
SET status_terkini = $2
FROM jadwal_sesi js
WHERE p.id = $1
  AND p.jadwal_sesi_id = js.id
  AND (
    (sqlc.narg('kelurahan_id')::smallint IS NOT NULL AND js.lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
    OR
    (sqlc.narg('kelurahan_id')::smallint IS NULL AND js.lokasi_kelurahan_id IS NULL)
  );

-- name: ListPendudukAdmin :many
SELECT 
    p.nik,
    p.nama_lengkap,
    p.jenis_kelamin,
    p.alamat,
    p.email,
    p.no_hp,
    k.nama_kelurahan
FROM penduduk p
LEFT JOIN ref_kelurahan k ON p.kelurahan_id = k.id
WHERE 
    (sqlc.narg('search')::text IS NULL OR 
     p.nik ILIKE '%' || sqlc.narg('search') || '%' OR 
     p.nama_lengkap ILIKE '%' || sqlc.narg('search') || '%' OR
     p.alamat ILIKE '%' || sqlc.narg('search') || '%')
    AND (sqlc.narg('kelurahan_id')::smallint IS NULL OR p.kelurahan_id = sqlc.narg('kelurahan_id'))
ORDER BY p.created_at DESC;

-- name: GetPendudukStatsAdmin :one
SELECT 
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE jenis_kelamin = 'LAKI_LAKI') AS laki_laki,
    COUNT(*) FILTER (WHERE jenis_kelamin = 'PEREMPUAN') AS perempuan,
    COUNT(*) AS wajib_ktp
FROM penduduk p
WHERE (sqlc.narg('kelurahan_id')::smallint IS NULL OR p.kelurahan_id = sqlc.narg('kelurahan_id'));

-- name: GetDokumenByPermohonan :many
SELECT 
    id,
    file_path,
    jenis_dokumen,
    uploaded_at
FROM dokumen_syarat
WHERE permohonan_id = $1;

-- name: ListTodayJadwal :many
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
WHERE js.tanggal = CURRENT_DATE
  AND (
    (sqlc.narg('kelurahan_id')::smallint IS NOT NULL AND js.lokasi_kelurahan_id = sqlc.narg('kelurahan_id'))
    OR
    (sqlc.narg('kelurahan_id')::smallint IS NULL AND js.lokasi_kelurahan_id IS NULL)
  )
ORDER BY js.jam_mulai;
