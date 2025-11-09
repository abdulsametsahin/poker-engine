package models

type Command struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Event struct {
	Event   string      `json:"event"`
	TableID string      `json:"tableId"`
	Data    interface{} `json:"data,omitempty"`
}

type ActionRequiredEvent struct {
	PlayerID string `json:"playerId"`
	Deadline string `json:"deadline"`
}

type ActionTimeoutEvent struct {
	PlayerID   string `json:"playerId"`
	AutoAction string `json:"autoAction"`
}

type HandCompleteEvent struct {
	Winners []Winner `json:"winners"`
}

type BlindsIncreasedEvent struct {
	NewSmallBlind int `json:"newSmallBlind"`
	NewBigBlind   int `json:"newBigBlind"`
}
