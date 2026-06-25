// Package xc — JSON adapter types for Xtream Codes API quirks.
//
// XC providers return the same logical field with inconsistent JSON types
// (an ID as int or string, a rating as float or string, episodes as [] or {}).
// These Flexible* types implement custom (un)marshaling to absorb that
// inconsistency so the domain DTOs in models.go stay simple.
package xc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexibleID handles IDs that may be returned as either int or string.
// Many XC providers return IDs inconsistently across endpoints.
type FlexibleID struct {
	intVal    int
	stringVal string
	isInt     bool
}

// UnmarshalJSON implements json.Unmarshaler for FlexibleID.
// Handles numeric (123), float (8.1 -> truncated to 8), and string ("123") JSON values.
func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	// Try integer first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		f.intVal = i
		f.isInt = true
		return nil
	}

	// Try float (truncate to int) - some APIs return floats for IDs
	var fl float64
	if err := json.Unmarshal(data, &fl); err == nil {
		f.intVal = int(fl)
		f.isInt = true
		return nil
	}

	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		f.stringVal = s
		f.isInt = false
		return nil
	}

	return fmt.Errorf("FlexibleID: cannot unmarshal %s", string(data))
}

// MarshalJSON implements json.Marshaler for FlexibleID.
func (f FlexibleID) MarshalJSON() ([]byte, error) {
	if f.isInt {
		return json.Marshal(f.intVal)
	}
	return json.Marshal(f.stringVal)
}

// String returns the ID as a string regardless of original type.
func (f FlexibleID) String() string {
	if f.isInt {
		return strconv.Itoa(f.intVal)
	}
	return f.stringVal
}

// Int returns the ID as an integer. If stored as string, attempts conversion.
func (f FlexibleID) Int() (int, error) {
	if f.isInt {
		return f.intVal, nil
	}
	return strconv.Atoi(f.stringVal)
}

// IsZero returns true if the ID is unset or zero.
func (f FlexibleID) IsZero() bool {
	if f.isInt {
		return f.intVal == 0
	}
	return f.stringVal == ""
}

// NewFlexibleID creates a FlexibleID from an integer.
func NewFlexibleID(id int) FlexibleID {
	return FlexibleID{intVal: id, isInt: true}
}

// NewFlexibleIDFromString creates a FlexibleID from a string.
func NewFlexibleIDFromString(id string) FlexibleID {
	return FlexibleID{stringVal: id, isInt: false}
}

// FlexibleFloat handles floats that may be returned as either float or string.
// XC providers often return ratings as strings like "4.5" instead of 4.5.
type FlexibleFloat float64

// UnmarshalJSON implements json.Unmarshaler for FlexibleFloat.
func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	// Try float first
	var fl float64
	if err := json.Unmarshal(data, &fl); err == nil {
		*f = FlexibleFloat(fl)
		return nil
	}

	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = 0
			return nil
		}
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			*f = 0
			return nil // Don't fail on unparseable strings
		}
		*f = FlexibleFloat(val)
		return nil
	}

	return fmt.Errorf("FlexibleFloat: cannot unmarshal %s", string(data))
}

// Float64 returns the value as float64.
func (f FlexibleFloat) Float64() float64 {
	return float64(f)
}

// FlexibleEpisodes handles episodes that may be returned as empty array or map.
// XC APIs return "episodes": [] when empty and "episodes": {"1": [...]} when populated.
type FlexibleEpisodes map[string][]Episode

// UnmarshalJSON implements json.Unmarshaler for FlexibleEpisodes.
func (f *FlexibleEpisodes) UnmarshalJSON(data []byte) error {
	// Check for empty array (common XC API behavior)
	if string(data) == "[]" || string(data) == "null" {
		*f = make(FlexibleEpisodes)
		return nil
	}

	// Try as map with string keys (standard format)
	var m map[string][]Episode
	if err := json.Unmarshal(data, &m); err == nil {
		*f = FlexibleEpisodes(m)
		return nil
	}

	// Try as map with integer keys (some providers use numeric keys)
	var intMap map[int][]Episode
	if err := json.Unmarshal(data, &intMap); err == nil {
		*f = make(FlexibleEpisodes)
		for k, v := range intMap {
			(*f)[strconv.Itoa(k)] = v
		}
		return nil
	}

	// Initialize empty if all else fails
	*f = make(FlexibleEpisodes)
	return nil
}

// FlexibleSeasons handles seasons that may be returned as array, false, or null.
// XC APIs sometimes return "seasons": false when no seasons exist.
type FlexibleSeasons []SeasonInfo

// UnmarshalJSON implements json.Unmarshaler for FlexibleSeasons.
func (f *FlexibleSeasons) UnmarshalJSON(data []byte) error {
	// Check for false/null (common XC API behavior)
	s := string(data)
	if s == "false" || s == "null" || s == "[]" {
		*f = make(FlexibleSeasons, 0)
		return nil
	}

	// Try as array
	var arr []SeasonInfo
	if err := json.Unmarshal(data, &arr); err == nil {
		*f = arr
		return nil
	}

	// Initialize empty if all else fails
	*f = make(FlexibleSeasons, 0)
	return nil
}

// FlexibleSeriesInfo handles info that may be returned as object, false, or null.
type FlexibleSeriesInfo struct {
	SeriesDetails
}

// UnmarshalJSON implements json.Unmarshaler for FlexibleSeriesInfo.
func (f *FlexibleSeriesInfo) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "false" || s == "null" || s == "[]" || s == "{}" {
		return nil
	}

	var details SeriesDetails
	if err := json.Unmarshal(data, &details); err != nil {
		return nil // Don't fail, just leave empty
	}
	f.SeriesDetails = details
	return nil
}

// FlexibleStringArr handles arrays that may be returned as single string, array, or false/null.
type FlexibleStringArr []string

// UnmarshalJSON implements json.Unmarshaler for FlexibleStringArr.
func (f *FlexibleStringArr) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "false" || s == "null" || s == "[]" {
		*f = make(FlexibleStringArr, 0)
		return nil
	}

	// Try as array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*f = arr
		return nil
	}

	// Try as single string
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single != "" {
			*f = []string{single}
		} else {
			*f = make(FlexibleStringArr, 0)
		}
		return nil
	}

	*f = make(FlexibleStringArr, 0)
	return nil
}
