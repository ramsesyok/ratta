package mode

import "ratta/internal/domain/issue"

// CanTransitionStatus は DD-DATA-003/F-004 の遷移許可を判定する。
func CanTransitionStatus(current issue.Status, next issue.Status, mode Mode) bool {
	if !current.IsValid() || !next.IsValid() {
		return false
	}
	if current.IsEndState() {
		return false
	}

	switch mode {
	case ModeContractor:
		return true
	case ModeVendor:
		if next == issue.StatusClosed || next == issue.StatusRejected {
			return false
		}
		return true
	default:
		return false
	}
}
