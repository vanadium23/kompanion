package entity

// Progress -.
type Progress struct {
	Document       string  `json:"document"`
	Percentage     float64 `json:"percentage"`
	Progress       string  `json:"progress"`
	Device         string  `json:"device"`
	DeviceID       string  `json:"device_id"`
	Timestamp      int64   `json:"timestamp"`
	AuthDeviceName string
}
