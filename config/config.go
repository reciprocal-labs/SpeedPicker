package config

import (
	"encoding/json"
	"os"
)

type LockConfig struct {
	Pin float64 `json:"pin"`
	SolvedState   float64   `json:"switch_state_solved"`
	Name string `json:"human_name"`
}

type BoardConfig struct {
	Locks []LockConfig `json:"locks"`
	LockDebounceTimeSeconds float64 `json:"lock_debounce_time_seconds"`
	StartButtonPin float64 `json:"start_button_pin"`
	ResetButtonPin float64 `json:"reset_button_pin"`
	StatusLedPin float64 `json:"status_led_pin"`
	HttpAddr string `json:"http_addr"`
}

func Load(filepath string) (*BoardConfig, error) {
	cfgFile, err := os.Open(filepath)
	if err != nil {
		return &BoardConfig{}, err
	}

	defer cfgFile.Close()

	parsedCfg := &BoardConfig{}

	jsonParser := json.NewDecoder(cfgFile)
	if err := jsonParser.Decode(&parsedCfg); err != nil {
		return &BoardConfig{}, err
	}

	return parsedCfg, nil
}

// String enables a BoardConfig to be serialized as a JSON string
func (b *BoardConfig) String() string {
	bytes, err := json.Marshal(b)
	if err != nil {
		return ""
	}

	return string(bytes)
}