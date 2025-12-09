-- +goose Up
-- +goose StatementBegin
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
        COALESCE(k.kode_area, 'KEC'), -- Handle NULL location
        TO_CHAR(j.tanggal, 'DD'),
        j.kuota_maksimal,
        j.kuota_terisi
    INTO v_kode_area, v_tanggal_day, v_kuota_max, v_kuota_now
    FROM jadwal_sesi j
    LEFT JOIN ref_kelurahan k ON j.lokasi_kelurahan_id = k.id -- LEFT JOIN
    WHERE j.id = NEW.jadwal_sesi_id
    FOR UPDATE OF j; -- Explicitly lock ONLY jadwal_sesi table

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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
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
        COALESCE(k.kode_area, 'KEC'), -- Handle NULL location
        TO_CHAR(j.tanggal, 'DD'),
        j.kuota_maksimal,
        j.kuota_terisi
    INTO v_kode_area, v_tanggal_day, v_kuota_max, v_kuota_now
    FROM jadwal_sesi j
    LEFT JOIN ref_kelurahan k ON j.lokasi_kelurahan_id = k.id -- LEFT JOIN
    WHERE j.id = NEW.jadwal_sesi_id
    FOR UPDATE; -- This was the problematic line

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
-- +goose StatementEnd
