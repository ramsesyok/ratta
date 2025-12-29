package mode

import (
	"ratta/internal/domain/issue"
	"testing"
)

func TestCanTransitionStatus_ContractorAllowsAllOpenMoves(t *testing.T) {
	// Contractor は終状態でなければ全てのステータスへ遷移できることを確認する。
	for _, next := range []issue.Status{
		issue.StatusOpen,
		issue.StatusWorking,
		issue.StatusInquiry,
		issue.StatusHold,
		issue.StatusFeedback,
		issue.StatusResolved,
		issue.StatusClosed,
		issue.StatusRejected,
	} {
		if !CanTransitionStatus(issue.StatusOpen, next, ModeContractor) {
			t.Fatalf("expected contractor to allow %s", next)
		}
	}
}

func TestCanTransitionStatus_VendorRejectsClosedAndRejected(t *testing.T) {
	// Vendor は Closed/Rejected への遷移を禁止されることを確認する。
	if CanTransitionStatus(issue.StatusOpen, issue.StatusClosed, ModeVendor) {
		t.Fatal("expected vendor to reject Closed")
	}
	if CanTransitionStatus(issue.StatusOpen, issue.StatusRejected, ModeVendor) {
		t.Fatal("expected vendor to reject Rejected")
	}
	if !CanTransitionStatus(issue.StatusOpen, issue.StatusResolved, ModeVendor) {
		t.Fatal("expected vendor to allow Resolved")
	}
}

func TestCanTransitionStatus_EndStateIsLocked(t *testing.T) {
	// 終状態からの遷移はモードに関係なく拒否されることを確認する。
	if CanTransitionStatus(issue.StatusClosed, issue.StatusOpen, ModeContractor) {
		t.Fatal("expected closed to be locked")
	}
	if CanTransitionStatus(issue.StatusRejected, issue.StatusOpen, ModeVendor) {
		t.Fatal("expected rejected to be locked")
	}
}
