package models

import (
	"errors"
	"fmt"
)

type StationType string

const (
	StationSolar   StationType = "SOLAR"
	StationWind    StationType = "WIND"
	StationBattery StationType = "BATTERY"
)

type GridStatus string

const (
	StatusNormal   GridStatus = "NORMAL"
	StatusWarning  GridStatus = "WARNING"
	StatusCritical GridStatus = "CRITICAL"
)

type ConsumerType string

const (
	ConsumerResidential ConsumerType = "RESIDENTIAL"
	ConsumerCommercial  ConsumerType = "COMMERCIAL"
	ConsumerIndustrial  ConsumerType = "INDUSTRIAL"
)

var (
	ErrInvalidStationID     = errors.New("invalid station id")
	ErrEmptyStationName     = errors.New("station name cannot be empty")
	ErrInvalidCapacity      = errors.New("capacity must be greater than zero")
	ErrInvalidStationType   = errors.New("invalid station type")
	ErrStationNotFound      = errors.New("station not found")
	ErrDuplicateStationID   = errors.New("station id already exists")
	ErrInvalidConsumerID    = errors.New("invalid consumer id")
	ErrEmptyConsumerName    = errors.New("consumer name cannot be empty")
	ErrInvalidConsumerType  = errors.New("invalid consumer type")
	ErrInvalidContractedMW  = errors.New("contracted mw must be greater than zero")
	ErrConsumerNotFound     = errors.New("consumer not found")
	ErrDuplicateConsumerID  = errors.New("consumer id already exists")
	ErrInsufficientEnergy   = errors.New("insufficient energy in source station")
	ErrStationOverload      = errors.New("requested output exceeds station capacity")
	ErrInvalidAmount        = errors.New("amount must be greater than zero")
	ErrConsumerDisconnected = errors.New("consumer is disconnected from grid")
	ErrStationInactive      = errors.New("station is currently inactive")
	ErrHubStopped           = errors.New("websocket hub is stopped")
)

type ValidationError struct {
	Field  string
	Reason string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation failed on field '%s': %s", v.Field, v.Reason)
}
