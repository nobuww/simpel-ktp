package user

import "github.com/nobuww/simpel-ktp/ui/components"

// GetExampleStages returns example stages for demo/development purposes
// In production, this should fetch actual stage data from the database
func GetExampleStages() []components.TrackerStage {
	return []components.TrackerStage{
		{
			ID:               "1",
			Name:             "Pengajuan",
			Description:      "Formulir permohonan telah disubmit dan menunggu verifikasi dokumen",
			Status:           components.StatusCompleted,
			StartDate:        "01 Des 2024, 09:00",
			CompletedDate:    "01 Des 2024, 09:15",
			ResponsibleParty: "Sistem Otomatis",
		},
		{
			ID:               "2",
			Name:             "Verifikasi Dokumen",
			Description:      "Tim kelurahan memverifikasi kelengkapan dan keabsahan dokumen yang diupload",
			Status:           components.StatusCompleted,
			StartDate:        "01 Des 2024, 09:15",
			CompletedDate:    "02 Des 2024, 14:30",
			ResponsibleParty: "Petugas Kelurahan",
		},
		{
			ID:                  "3",
			Name:                "Proses Pencetakan",
			Description:         "Data Anda sedang diproses untuk pencetakan KTP di Disdukcapil",
			Status:              components.StatusInProgress,
			StartDate:           "02 Des 2024, 15:00",
			EstimatedCompletion: "07 Des 2024",
			ResponsibleParty:    "Disdukcapil Kota",
		},
		{
			ID:                  "4",
			Name:                "Siap Ambil",
			Description:         "KTP Anda telah selesai dicetak dan siap untuk diambil di lokasi yang ditentukan",
			Status:              components.StatusPending,
			EstimatedCompletion: "08 Des 2024",
			ResponsibleParty:    "Kelurahan",
			DocumentsRequired:   []string{"KK Asli", "Surat Permohonan Asli", "Bukti Booking"},
		},
		{
			ID:                  "5",
			Name:                "Selesai",
			Description:         "KTP telah diambil oleh pemohon",
			Status:              components.StatusPending,
			EstimatedCompletion: "09 Des 2024",
			ResponsibleParty:    "Pemohon",
		},
	}
}

// GetBlockedExampleStages returns example stages with a blocked status for demo
func GetBlockedExampleStages() []components.TrackerStage {
	return []components.TrackerStage{
		{
			ID:               "1",
			Name:             "Pengajuan",
			Description:      "Formulir permohonan telah disubmit",
			Status:           components.StatusCompleted,
			StartDate:        "01 Des 2024, 09:00",
			CompletedDate:    "01 Des 2024, 09:15",
			ResponsibleParty: "Sistem Otomatis",
		},
		{
			ID:                "2",
			Name:              "Verifikasi Dokumen",
			Description:       "Dokumen tidak lengkap atau tidak valid",
			Status:            components.StatusBlocked,
			StartDate:         "01 Des 2024, 09:15",
			Notes:             "Foto KK tidak terbaca dengan jelas. Silakan upload ulang dengan kualitas lebih baik.",
			ResponsibleParty:  "Pemohon",
			DocumentsRequired: []string{"Foto KK (kualitas tinggi)", "Foto diri dengan latar biru"},
		},
		{
			ID:          "3",
			Name:        "Proses Pencetakan",
			Description: "Data akan diproses setelah verifikasi selesai",
			Status:      components.StatusPending,
		},
		{
			ID:          "4",
			Name:        "Siap Ambil",
			Description: "KTP siap untuk diambil",
			Status:      components.StatusPending,
		},
		{
			ID:          "5",
			Name:        "Selesai",
			Description: "KTP telah diambil",
			Status:      components.StatusPending,
		},
	}
}
