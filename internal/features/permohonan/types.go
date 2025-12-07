package permohonan

// FormData contains user data to pre-fill forms
type FormData struct {
	NIK           string
	NamaLengkap   string
	JenisKelamin  string
	Alamat        string
	NoHP          string
	Email         string
	NamaKelurahan string
	Errors        map[string]string
}

// JadwalOption represents a jadwal option for the select box
type JadwalOption struct {
	ID         string
	Label      string
	KuotaSisa  int
	StatusSesi string
}

type LocationOption struct {
	ID    int16
	Label string
}
