package mode

// Mode は DD-BE-003 のモード種別を表す。
type Mode string

const (
	ModeContractor Mode = "Contractor"
	ModeVendor     Mode = "Vendor"
)
