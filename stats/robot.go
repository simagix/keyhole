package stats

// Task -
type Task struct {
	For         string
	MinutesUsed int
}

// Robot -
type Robot struct {
	ID         string `json:"_id" bson:"_id"`
	ModelID    string `json:"modelId,omitempty" bson:"modelId,omitempty"`
	Notes      string
	BatteryPct float32 `json:"batteryPct,omitempty" bson:"batteryPct,omitempty"`
	Tasks      []Task
}
