package controllers

import "testing"

func TestParseTargetAppVersionID(t *testing.T) {
	patterns := []struct {
		s                 string
		expectedAppID     int
		expectedVersionID int
		expectedErr       string
	}{
		{
			s:                 "1_2",
			expectedAppID:     1,
			expectedVersionID: 2,
		},
		{
			s:           "",
			expectedErr: "malformed target app version id",
		},
		{
			s:           "A_2",
			expectedErr: "A is not appId",
		},
		{
			s:           "1_B",
			expectedErr: "B is not version",
		},
	}

	for i, pattern := range patterns {
		appID, versionID, err := parseTargetAppVersionID(pattern.s)
		if err != nil {
			if err.Error() != pattern.expectedErr {
				t.Fatalf("#%d: unexpected error: %v", i, err)
			}
			continue
		}
		if appID != pattern.expectedAppID {
			t.Fatalf("#%d: appID is wrong: %v", i, appID)
		}
		if versionID != pattern.expectedVersionID {
			t.Fatalf("#%d: versionID is wrong: %v", i, versionID)
		}
	}
}
