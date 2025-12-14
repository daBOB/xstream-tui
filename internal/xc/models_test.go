package xc

import (
	"encoding/json"
	"testing"
)

func TestFlexibleID_UnmarshalJSON_Int(t *testing.T) {
	input := `123`
	var fid FlexibleID
	if err := json.Unmarshal([]byte(input), &fid); err != nil {
		t.Fatalf("unmarshal int: %v", err)
	}

	if !fid.isInt {
		t.Error("expected isInt to be true")
	}
	if fid.String() != "123" {
		t.Errorf("String() = %q, want %q", fid.String(), "123")
	}
	if val, err := fid.Int(); err != nil || val != 123 {
		t.Errorf("Int() = %d, %v, want 123, nil", val, err)
	}
}

func TestFlexibleID_UnmarshalJSON_String(t *testing.T) {
	input := `"456"`
	var fid FlexibleID
	if err := json.Unmarshal([]byte(input), &fid); err != nil {
		t.Fatalf("unmarshal string: %v", err)
	}

	if fid.isInt {
		t.Error("expected isInt to be false")
	}
	if fid.String() != "456" {
		t.Errorf("String() = %q, want %q", fid.String(), "456")
	}
	if val, err := fid.Int(); err != nil || val != 456 {
		t.Errorf("Int() = %d, %v, want 456, nil", val, err)
	}
}

func TestFlexibleID_UnmarshalJSON_InvalidType(t *testing.T) {
	input := `true`
	var fid FlexibleID
	err := json.Unmarshal([]byte(input), &fid)
	if err == nil {
		t.Error("expected error for boolean input")
	}
}

func TestFlexibleID_MarshalJSON_Int(t *testing.T) {
	fid := NewFlexibleID(789)
	data, err := json.Marshal(fid)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != "789" {
		t.Errorf("marshal result = %s, want 789", string(data))
	}
}

func TestFlexibleID_MarshalJSON_String(t *testing.T) {
	fid := NewFlexibleIDFromString("abc")
	data, err := json.Marshal(fid)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != `"abc"` {
		t.Errorf("marshal result = %s, want \"abc\"", string(data))
	}
}

func TestFlexibleID_IsZero(t *testing.T) {
	tests := []struct {
		name string
		fid  FlexibleID
		want bool
	}{
		{"zero int", NewFlexibleID(0), true},
		{"non-zero int", NewFlexibleID(1), false},
		{"empty string", NewFlexibleIDFromString(""), true},
		{"non-empty string", NewFlexibleIDFromString("a"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fid.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlexibleID_Int_StringConversion(t *testing.T) {
	// Valid numeric string
	fid := NewFlexibleIDFromString("999")
	val, err := fid.Int()
	if err != nil {
		t.Errorf("Int() error = %v", err)
	}
	if val != 999 {
		t.Errorf("Int() = %d, want 999", val)
	}

	// Invalid string
	fid2 := NewFlexibleIDFromString("not-a-number")
	_, err = fid2.Int()
	if err == nil {
		t.Error("expected error for non-numeric string")
	}
}

func TestCategory_UnmarshalJSON(t *testing.T) {
	// Test with numeric category_id
	jsonNumeric := `{"category_id": 123, "category_name": "Sports", "parent_id": 0}`
	var cat1 Category
	if err := json.Unmarshal([]byte(jsonNumeric), &cat1); err != nil {
		t.Fatalf("unmarshal numeric: %v", err)
	}
	if cat1.Name != "Sports" {
		t.Errorf("Name = %q, want %q", cat1.Name, "Sports")
	}
	if cat1.ID.String() != "123" {
		t.Errorf("ID = %q, want %q", cat1.ID.String(), "123")
	}

	// Test with string category_id
	jsonString := `{"category_id": "456", "category_name": "Movies", "parent_id": "1"}`
	var cat2 Category
	if err := json.Unmarshal([]byte(jsonString), &cat2); err != nil {
		t.Fatalf("unmarshal string: %v", err)
	}
	if cat2.Name != "Movies" {
		t.Errorf("Name = %q, want %q", cat2.Name, "Movies")
	}
	if cat2.ID.String() != "456" {
		t.Errorf("ID = %q, want %q", cat2.ID.String(), "456")
	}
}

func TestLiveStream_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"stream_id": 12345,
		"name": "ESPN HD",
		"stream_icon": "http://example.com/icon.png",
		"category_id": "7",
		"epg_channel_id": "espn.us",
		"num": "1",
		"added": 1609459200,
		"is_adult": 0,
		"tv_archive": 1
	}`

	var stream LiveStream
	if err := json.Unmarshal([]byte(jsonData), &stream); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if stream.ID.String() != "12345" {
		t.Errorf("ID = %q, want %q", stream.ID.String(), "12345")
	}
	if stream.Name != "ESPN HD" {
		t.Errorf("Name = %q, want %q", stream.Name, "ESPN HD")
	}
	if stream.CategoryID.String() != "7" {
		t.Errorf("CategoryID = %q, want %q", stream.CategoryID.String(), "7")
	}
	if stream.EPGChannelID != "espn.us" {
		t.Errorf("EPGChannelID = %q, want %q", stream.EPGChannelID, "espn.us")
	}
}

func TestVODStream_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"stream_id": "98765",
		"name": "The Matrix",
		"stream_icon": "http://example.com/matrix.jpg",
		"category_id": 3,
		"container_extension": "mkv",
		"rating": "8.5",
		"rating_5based": 4.25
	}`

	var vod VODStream
	if err := json.Unmarshal([]byte(jsonData), &vod); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if vod.ID.String() != "98765" {
		t.Errorf("ID = %q, want %q", vod.ID.String(), "98765")
	}
	if vod.Name != "The Matrix" {
		t.Errorf("Name = %q, want %q", vod.Name, "The Matrix")
	}
	if vod.Container != "mkv" {
		t.Errorf("Container = %q, want %q", vod.Container, "mkv")
	}
	if vod.Rating5Based != 4.25 {
		t.Errorf("Rating5Based = %f, want 4.25", vod.Rating5Based)
	}
}

func TestUserInfo_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"username": "testuser",
		"password": "testpass",
		"status": "Active",
		"active_cons": "1",
		"max_connections": 2,
		"exp_date": "1735689600",
		"is_trial": "0"
	}`

	var user UserInfo
	if err := json.Unmarshal([]byte(jsonData), &user); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Username = %q, want %q", user.Username, "testuser")
	}
	if user.Status != "Active" {
		t.Errorf("Status = %q, want %q", user.Status, "Active")
	}
	if user.MaxConnections.String() != "2" {
		t.Errorf("MaxConnections = %q, want %q", user.MaxConnections.String(), "2")
	}
}

func TestAuthResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"user_info": {
			"username": "demo",
			"status": "Active",
			"max_connections": "1"
		},
		"server_info": {
			"url": "example.com",
			"port": "80",
			"https_port": "443",
			"server_protocol": "http"
		}
	}`

	var auth AuthResponse
	if err := json.Unmarshal([]byte(jsonData), &auth); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if auth.UserInfo.Username != "demo" {
		t.Errorf("UserInfo.Username = %q, want %q", auth.UserInfo.Username, "demo")
	}
	if auth.ServerInfo.URL != "example.com" {
		t.Errorf("ServerInfo.URL = %q, want %q", auth.ServerInfo.URL, "example.com")
	}
}
