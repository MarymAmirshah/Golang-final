package service

import (
	"errors"
	"sync"
	"time"

	"GoPower/models"
	"GoPower/repository"

	"gorm.io/gorm"
)

type GridService struct {
	db           *gorm.DB
	stationRepo  repository.StationRepository
	consumerRepo repository.ConsumerRepository
	mu           sync.Mutex
}

func NewGridService(db *gorm.DB, stationRepo repository.StationRepository, consumerRepo repository.ConsumerRepository) *GridService {
	return &GridService{
		db:           db,
		stationRepo:  stationRepo,
		consumerRepo: consumerRepo,
	}
}

// RecordGeneration records the current instantaneous output of an active
// power station.
func (s *GridService) RecordGeneration(stationID string, outputMW float64) error {
	if outputMW < 0 {
		return models.ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return err
	}

	if !station.IsActive {
		return models.ErrStationInactive
	}

	if outputMW > station.CapacityMW {
		return models.ErrStationOverload
	}

	station.CurrentOutput = outputMW
	return s.stationRepo.Update(station)
}

// ChargeBattery injects surplus generated energy into a battery storage
// station.
func (s *GridService) ChargeBattery(stationID string, amountMWh float64) error {
	if amountMWh <= 0 {
		return models.ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return err
	}

	if station.Type != models.StationBattery {
		return models.ErrInvalidStationType
	}
	if !station.IsActive {
		return models.ErrStationInactive
	}

	if station.StoredEnergy+amountMWh > station.MaxStorage {
		return models.ErrStationOverload
	}

	station.StoredEnergy += amountMWh
	return s.stationRepo.Update(station)
}

// DispatchEnergy atomically transfers a load of energy from a station to a
// consumer, recording the transaction as a DispatchRecord.
func (s *GridService) DispatchEnergy(stationID, consumerID string, amountMW float64) (*models.DispatchRecord, error) {
	if amountMW <= 0 {
		return nil, models.ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var record models.DispatchRecord

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var station models.PowerStation
		if err := tx.First(&station, "id = ?", stationID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.ErrStationNotFound
			}
			return err
		}
		if !station.IsActive {
			return models.ErrStationInactive
		}

		var consumer models.ConsumerMeter
		if err := tx.First(&consumer, "id = ?", consumerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.ErrConsumerNotFound
			}
			return err
		}
		if !consumer.IsConnected {
			return models.ErrConsumerDisconnected
		}

		if station.Type == models.StationBattery {
			if station.StoredEnergy < amountMW {
				return models.ErrInsufficientEnergy
			}
		} else {
			if station.CurrentOutput < amountMW {
				return models.ErrInsufficientEnergy
			}
		}

		if consumer.CurrentLoad+amountMW > consumer.ContractedMW {
			return models.ErrStationOverload
		}

		if station.Type == models.StationBattery {
			station.StoredEnergy -= amountMW
		} else {
			station.CurrentOutput -= amountMW
		}
		consumer.CurrentLoad += amountMW

		if err := tx.Save(&station).Error; err != nil {
			return err
		}
		if err := tx.Save(&consumer).Error; err != nil {
			return err
		}

		record = models.DispatchRecord{
			StationID:   stationID,
			ConsumerID:  consumerID,
			AmountMW:    amountMW,
			DeliveredAt: time.Now(),
		}
		return tx.Create(&record).Error
	})

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// GetGridSummary computes an instantaneous snapshot of the whole
// microgrid's generation, demand, and balance.
func (s *GridService) GetGridSummary() (*models.GridSummary, error) {
	stations, err := s.stationRepo.ListAll()
	if err != nil {
		return nil, err
	}

	consumers, err := s.consumerRepo.ListAll()
	if err != nil {
		return nil, err
	}

	summary := &models.GridSummary{
		StationBreakdown: make(map[models.StationType]int),
	}

	for _, st := range stations {
		if !st.IsActive {
			continue
		}
		summary.TotalStations++
		summary.TotalCapacityMW += st.CapacityMW
		summary.TotalGenerationMW += st.CurrentOutput
		summary.StationBreakdown[st.Type]++
	}

	for _, cm := range consumers {
		if !cm.IsConnected {
			continue
		}
		summary.TotalConsumers++
		summary.TotalDemandMW += cm.CurrentLoad
	}

	summary.NetBalanceMW = summary.TotalGenerationMW - summary.TotalDemandMW

	switch {
	case summary.NetBalanceMW >= 0:
		summary.Status = models.StatusNormal
	case summary.NetBalanceMW >= -50:
		summary.Status = models.StatusWarning
	default:
		summary.Status = models.StatusCritical
	}

	return summary, nil
}
