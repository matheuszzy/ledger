package service

import "testing"

func TestService_Reset(t *testing.T) {
	svc, store := newService()
	seed(t, store, "100", 10)
	seed(t, store, "300", 15)

	svc.Reset()

	assertMissing(t, store, "100")
	assertMissing(t, store, "300")
	if _, err := svc.Deposit("100", 1); err != nil {
		t.Fatalf("service unusable after Reset: %v", err)
	}
	assertBalance(t, store, "100", 1)
}
