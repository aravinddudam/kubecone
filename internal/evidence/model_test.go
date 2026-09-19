package evidence

import "testing"

func TestScanItemUnhealthy(t *testing.T) {
	cases := []struct {
		code, status string
		want         bool
	}{
		{CodeImagePull, "ImagePullBackOff", true},
		{CodeHealthy, "Ready", false},
		{CodeHealthy, "Running", false},
		{"", "NotReady", true},
		{"", "ImagePullBackOff", true},
		{"", "Pending", true},
		{"", "Ready", false},
		{"", "Succeeded", false},
		{"", "", false},
	}
	for _, tc := range cases {
		item := ScanItem{Code: tc.code, Status: tc.status}
		if got := item.Unhealthy(); got != tc.want {
			t.Fatalf("code=%q status=%q unhealthy=%v want=%v", tc.code, tc.status, got, tc.want)
		}
	}
}
