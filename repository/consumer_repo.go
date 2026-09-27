package repository

import (
	"errors"

	"GoPower/models"

	"gorm.io/gorm"
)

type ConsumerRepository interface {
	Create(consumer *models.ConsumerMeter) error
	GetByID(id string) (*models.ConsumerMeter, error)
	ListAll() ([]models.ConsumerMeter, error)
	Update(consumer *models.ConsumerMeter) error
	Delete(id string) error
}

type GormConsumerRepository struct {
	db *gorm.DB
}

func NewConsumerRepository(db *gorm.DB) *GormConsumerRepository {
	return &GormConsumerRepository{db: db}
}

func (r *GormConsumerRepository) Create(consumer *models.ConsumerMeter) error {
	if err := consumer.Validate(); err != nil {
		return err
	}

	var existing models.ConsumerMeter
	err := r.db.First(&existing, "id = ?", consumer.ID).Error
	switch {
	case err == nil:
		return models.ErrDuplicateConsumerID
	case errors.Is(err, gorm.ErrRecordNotFound):
		// no conflict, proceed
	default:
		return err
	}

	if err := r.db.Create(consumer).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return models.ErrDuplicateConsumerID
		}
		return err
	}

	return nil
}

func (r *GormConsumerRepository) GetByID(id string) (*models.ConsumerMeter, error) {
	if id == "" {
		return nil, models.ErrInvalidConsumerID
	}

	var consumer models.ConsumerMeter
	if err := r.db.First(&consumer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrConsumerNotFound
		}
		return nil, err
	}

	return &consumer, nil
}

func (r *GormConsumerRepository) ListAll() ([]models.ConsumerMeter, error) {
	var consumers []models.ConsumerMeter
	if err := r.db.Order("created_at ASC").Find(&consumers).Error; err != nil {
		return nil, err
	}
	return consumers, nil
}

func (r *GormConsumerRepository) Update(consumer *models.ConsumerMeter) error {
	if err := consumer.Validate(); err != nil {
		return err
	}

	var existing models.ConsumerMeter
	err := r.db.First(&existing, "id = ?", consumer.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ErrConsumerNotFound
		}
		return err
	}

	return r.db.Save(consumer).Error
}

func (r *GormConsumerRepository) Delete(id string) error {
	if id == "" {
		return models.ErrInvalidConsumerID
	}

	result := r.db.Delete(&models.ConsumerMeter{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrConsumerNotFound
	}

	return nil
}
