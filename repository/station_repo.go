package repository

import (
	"errors"

	"GoPower/models"

	"gorm.io/gorm"
)

type StationRepository interface {
	Create(station *models.PowerStation) error
	GetByID(id string) (*models.PowerStation, error)
	ListAll() ([]models.PowerStation, error)
	Update(station *models.PowerStation) error
	Delete(id string) error
	GetByType(stationType models.StationType) ([]models.PowerStation, error)
}

type GormStationRepository struct {
	db *gorm.DB
}

func NewStationRepository(db *gorm.DB) *GormStationRepository {
	return &GormStationRepository{db: db}
}

func (r *GormStationRepository) Create(station *models.PowerStation) error {
	if err := station.Validate(); err != nil {
		return err
	}

	var existing models.PowerStation
	err := r.db.First(&existing, "id = ?", station.ID).Error
	switch {
	case err == nil:
		return models.ErrDuplicateStationID
	case errors.Is(err, gorm.ErrRecordNotFound):
		// no conflict, proceed
	default:
		return err
	}

	if err := r.db.Create(station).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return models.ErrDuplicateStationID
		}
		return err
	}

	return nil
}

func (r *GormStationRepository) GetByID(id string) (*models.PowerStation, error) {
	if id == "" {
		return nil, models.ErrInvalidStationID
	}

	var station models.PowerStation
	if err := r.db.First(&station, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrStationNotFound
		}
		return nil, err
	}

	return &station, nil
}

func (r *GormStationRepository) ListAll() ([]models.PowerStation, error) {
	var stations []models.PowerStation
	if err := r.db.Order("created_at ASC").Find(&stations).Error; err != nil {
		return nil, err
	}
	return stations, nil
}

func (r *GormStationRepository) Update(station *models.PowerStation) error {
	if err := station.Validate(); err != nil {
		return err
	}

	var existing models.PowerStation
	err := r.db.First(&existing, "id = ?", station.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ErrStationNotFound
		}
		return err
	}

	return r.db.Save(station).Error
}

func (r *GormStationRepository) Delete(id string) error {
	if id == "" {
		return models.ErrInvalidStationID
	}

	result := r.db.Delete(&models.PowerStation{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrStationNotFound
	}

	return nil
}

func (r *GormStationRepository) GetByType(stationType models.StationType) ([]models.PowerStation, error) {
	var stations []models.PowerStation
	if err := r.db.Where("type = ?", stationType).Order("created_at ASC").Find(&stations).Error; err != nil {
		return nil, err
	}
	return stations, nil
}
