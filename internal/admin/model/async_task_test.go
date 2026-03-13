package model

import "testing"

func TestResolveAsyncTaskBusinessOutcomeForSupportedChannelModelTest(t *testing.T) {
	status, message, ok := ResolveAsyncTaskBusinessOutcome(
		AsyncTaskTypeChannelModelTest,
		`{"status":"supported","supported":true,"message":"ok"}`,
	)
	if !ok {
		t.Fatalf("ResolveAsyncTaskBusinessOutcome ok = false, want true")
	}
	if status != AsyncTaskStatusSucceeded {
		t.Fatalf("ResolveAsyncTaskBusinessOutcome status = %q, want %q", status, AsyncTaskStatusSucceeded)
	}
	if message != "" {
		t.Fatalf("ResolveAsyncTaskBusinessOutcome message = %q, want empty", message)
	}
}

func TestResolveAsyncTaskBusinessOutcomeForUnsupportedChannelModelTest(t *testing.T) {
	status, message, ok := ResolveAsyncTaskBusinessOutcome(
		AsyncTaskTypeChannelModelTest,
		`{"status":"unsupported","supported":false,"message":"dial tcp timeout"}`,
	)
	if !ok {
		t.Fatalf("ResolveAsyncTaskBusinessOutcome ok = false, want true")
	}
	if status != AsyncTaskStatusFailed {
		t.Fatalf("ResolveAsyncTaskBusinessOutcome status = %q, want %q", status, AsyncTaskStatusFailed)
	}
	if message != "dial tcp timeout" {
		t.Fatalf("ResolveAsyncTaskBusinessOutcome message = %q, want %q", message, "dial tcp timeout")
	}
}

func TestNormalizeAsyncTaskRowAppliesBusinessFailureStatus(t *testing.T) {
	row := AsyncTask{
		Type:   AsyncTaskTypeChannelModelTest,
		Status: AsyncTaskStatusSucceeded,
		Result: `{"status":"unsupported","supported":false,"message":"request timeout"}`,
	}

	normalizeAsyncTaskRow(&row)

	if row.Status != AsyncTaskStatusFailed {
		t.Fatalf("normalizeAsyncTaskRow status = %q, want %q", row.Status, AsyncTaskStatusFailed)
	}
	if row.ErrorMessage != "request timeout" {
		t.Fatalf("normalizeAsyncTaskRow error_message = %q, want %q", row.ErrorMessage, "request timeout")
	}
}
