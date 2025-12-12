package migration

import (
	"errors"
	"fmt"

	"github.com/231031/pethealth-backend/internal/applogger"
	"github.com/231031/pethealth-backend/internal/model"
	"gorm.io/gorm"
)

var (
	migrateLevel = "MIGRATION"
)

type MigrationManager struct {
	DB     *gorm.DB
	models []interface{}
}

func NewMigrationManager(db *gorm.DB) *MigrationManager {
	models := []interface{}{
		&model.User{},
		&model.Payment{},
		&model.Pet{},
		&model.PetDetail{},
		&model.Food{},
		&model.CupFoodPet{},
		&model.PetFoodPlan{},
		&model.FoodPetFoodPlan{},
		&model.PetFoodPlanDetail{},
		&model.PetFoodPlanHistory{},
		&model.PetCalendar{},
		&model.PetActivityCalendar{},
		&model.PetSkinImage{},
	}

	return &MigrationManager{
		DB:     db,
		models: models,
	}
}

func (m *MigrationManager) MigrateToDB() error {
	var modelsToMigrate []interface{}
	for i := range m.models {
		if m.DB.Migrator().HasTable(m.models[i]) {
			continue
		}

		modelsToMigrate = append(modelsToMigrate, m.models[i])
	}
	applogger.LogInfo(fmt.Sprintln("migrating new model to db : ", len(modelsToMigrate)), migrateLevel)

	if err := m.DB.AutoMigrate(modelsToMigrate...); err != nil {
		applogger.LogError(fmt.Sprintln("failed to migrate to db : ", err), migrateLevel)
		return errors.New("failed to migrate to db")
	}
	return nil
}

func (m *MigrationManager) DropAllTables() error {
	if m.DB == nil {
		applogger.LogError("database connection is nil", migrateLevel)
		return errors.New("database connection is nil")
	}

	err := m.DB.Migrator().DropTable(m.models...)
	if err != nil {
		applogger.LogError(fmt.Sprintln("failed to drop tables :", err), migrateLevel)
		return errors.New("failed to drop tables")
	}
	return nil
}

func (m *MigrationManager) ResetDB() error {
	err := m.DropAllTables()
	if err != nil {
		return err
	}
	err = m.MigrateToDB()
	if err != nil {
		return err
	}
	return nil
}
