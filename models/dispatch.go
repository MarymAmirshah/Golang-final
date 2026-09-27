package models

import "time"

type DispatchRecord struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	StationID   string    `gorm:"not null;size:64;index" json:"station_id"`
	ConsumerID  string    `gorm:"not null;size:64;index" json:"consumer_id"`
	AmountMW    float64   `gorm:"not null" json:"amount_mw"`
	DeliveredAt time.Time `gorm:"not null" json:"delivered_at"`
}

type GridSummary struct {
	TotalStations     int                 `json:"total_stations"`
	TotalGenerationMW float64             `json:"total_generation_mw"`
	TotalCapacityMW   float64             `json:"total_capacity_mw"`
	TotalConsumers    int                 `json:"total_consumers"`
	TotalDemandMW     float64             `json:"total_demand_mw"`
	NetBalanceMW      float64             `json:"net_balance_mw"`
	Status            GridStatus          `json:"status"`
	StationBreakdown  map[StationType]int `json:"station_breakdown"`
}
