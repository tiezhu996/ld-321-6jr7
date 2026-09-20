package repository

import (
	"errors"
	"fmt"

	"github.com/agridispatch/agridispatch/internal/model"
	"gorm.io/gorm"
)

// FindDriverByName 按姓名查找驾驶员。
func (r *DashboardRepository) FindDriverByName(name string) (*model.Driver, error) {
	var d model.Driver
	err := r.db.First(&d, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find driver by name: %w", err)
	}
	return &d, nil
}

// FindActiveRemindersByMachine 查询农机当前生效的全部保养提醒。
func (r *DashboardRepository) FindActiveRemindersByMachine(machineCode string) ([]model.MaintenanceReminder, error) {
	var reminders []model.MaintenanceReminder
	if err := r.db.
		Where("machine_code = ?", machineCode).
		Find(&reminders).Error; err != nil {
		return nil, fmt.Errorf("find reminders by machine: %w", err)
	}
	return reminders, nil
}
