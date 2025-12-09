-- name: GetPermohonanByNIK :many
SELECT 
    p.id,
    p.kode_booking,
    p.jenis_permohonan,
    p.status_terkini,
    p.nomor_antrian_sesi as nomor_antrian,
    p.created_at,
    js.tanggal as jadwal_tanggal,
    js.jam_mulai as jadwal_jam_mulai,
    js.jam_selesai as jadwal_jam_selesai,
    k.nama_kelurahan as lokasi_kelurahan
FROM permohonan p
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
LEFT JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
WHERE p.nik = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPermohonanByNIK :one
SELECT 
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE status_terkini = 'VERIFIKASI') as verifikasi,
    COUNT(*) FILTER (WHERE status_terkini = 'PROSES') as proses,
    COUNT(*) FILTER (WHERE status_terkini = 'SIAP_AMBIL') as siap_ambil,
    COUNT(*) FILTER (WHERE status_terkini = 'SELESAI') as selesai,
    COUNT(*) FILTER (WHERE status_terkini = 'DITOLAK') as ditolak
FROM permohonan 
WHERE nik = $1;

-- name: GetPendudukProfile :one
SELECT 
    p.nik,
    p.nama_lengkap,
    p.email,
    p.alamat,
    p.no_hp,
    p.jenis_kelamin,
    p.created_at,
    k.nama_kelurahan
FROM penduduk p
LEFT JOIN ref_kelurahan k ON p.kelurahan_id = k.id
WHERE p.nik = $1;

-- name: GetPermohonanByKodeBooking :one
SELECT 
    p.id,
    p.kode_booking,
    p.jenis_permohonan,
    p.status_terkini,
    p.created_at as tanggal_daftar,
    p.nomor_antrian_sesi as nomor_antrian,
    js.tanggal as jadwal_tanggal,
    js.jam_mulai as jadwal_jam_mulai,
    js.jam_selesai as jadwal_jam_selesai,
    k.nama_kelurahan as lokasi_kelurahan,
    pd.nama_lengkap
FROM permohonan p
LEFT JOIN penduduk pd ON p.nik = pd.nik
LEFT JOIN jadwal_sesi js ON p.jadwal_sesi_id = js.id
LEFT JOIN ref_kelurahan k ON js.lokasi_kelurahan_id = k.id
WHERE p.kode_booking = $1;
