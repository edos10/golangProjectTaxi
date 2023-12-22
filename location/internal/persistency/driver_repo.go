package persistency

import (
	"context"
	"go.uber.org/zap"
	
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/google/uuid"
	"github.com/location/internal/model"
)

type DriverRepo interface {
	GetDrivers(ctx context.Context) (*[]model.Driver, error)
	UpdateDriverLocation(ctx context.Context, driverId uuid.UUID, location model.Location) error
	CheckDriverExistence(ctx context.Context, driverId uuid.UUID) (bool, error)
}

type driverRepo struct {
	db *pgxpool.Pool
	logger *zap.Logger
}

func (repo *driverRepo) GetDrivers(ctx context.Context) (*[]model.Driver, error) {
	query := "SELECT Id, Lat, Lng FROM drivers"

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		repo.logger.Error("Error with GetDrivers query", zap.Error(err))
		return nil, err
	}

	var drivers []model.Driver
	for rows.Next() {
		var driver model.Driver
		if err := rows.Scan(&driver.ID, &driver.Location.Lat, &driver.Location.Lng); err != nil {
			repo.logger.Error("Error with scanning drivers rows", zap.Error(err))
			return nil, err
		}
		drivers = append(drivers, driver)
	}
	if err := rows.Err(); err != nil {
		repo.logger.Error("Error with GetDrivers query", zap.Error(err))
		return nil, err
	}

	return &drivers, nil
}

func (repo *driverRepo) UpdateDriverLocation(ctx context.Context, driverId uuid.UUID, location model.Location) error {
	query := "UPDATE drivers SET Lat = $1, Lng = $2 WHERE ID = $3"

	_, err := repo.db.Exec(ctx, query, location.Lat, location.Lng, driverId)
	if err != nil {
		repo.logger.Error("Error with UpdateDriver query", zap.Error(err))
		return err
	}
	return nil
}

func (repo *driverRepo) CheckDriverExistence(ctx context.Context, driverId uuid.UUID) (bool, error) {
	drivers, err := repo.GetDrivers(ctx)
	if err != nil {
		repo.logger.Error("Error with Checking driver existence", zap.Error(err))
		return false, err
	}

	for _, driver := range *drivers {
		if driver.ID == driverId {
			return true, nil
		}
	}
	
	return false, nil
}


func NewDriverRepo(db *pgxpool.Pool, logger *zap.Logger) (DriverRepo, error) {
	repo := &driverRepo{
		db: db,
		logger: logger,
	}

	return repo, nil
}
