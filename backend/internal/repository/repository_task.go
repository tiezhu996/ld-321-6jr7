package repository

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/agridispatch/agridispatch/internal/constants"
	apperrors "github.com/agridispatch/agridispatch/internal/errors"
	"github.com/agridispatch/agridispatch/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TaskCompletionPlan 完工登记执行计划（业务层校验并计算油耗成本后下发）。
type TaskCompletionPlan struct {
	TaskID      string
	ActualHours float64
	FuelLiters  float64
	AreaMu      float64
	FuelCost    float64
	WorkDate    string
	Now         time.Time
}

// TaskCompletionResult 完工登记结果（用于响应与日志）。
type TaskCompletionResult struct {
	Record         *model.WorkRecord `json:"record"`
	TaskID         string            `json:"taskId"`
	TaskStatus     string            `json:"taskStatus"`
	MachineCode    string            `json:"machineCode"`
	MachineStatus  string            `json:"machineStatus"`
	RemainingHours float64           `json:"remainingHours"`
	MaintenanceDue bool              `json:"maintenanceDue"`
	DriverName     string            `json:"driverName"`
	DriverStatus   string            `json:"driverStatus"`
}

// TaskRepository 作业任务数据访问（派单 / 完工登记）。
type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Dispatch 派单事务：任务待派单 → 已派单，农机空闲 → 作业中，驾驶员 → 作业中。
// 维修中或作业中的农机、作业中的驾驶员均阻止派单。
func (r *TaskRepository) Dispatch(taskID string) (machineCode, driverName string, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		task, lockErr := lockTask(tx, taskID)
		if lockErr != nil {
			return lockErr
		}
		if task.Status == constants.TaskDispatched || task.Status == constants.TaskDone {
			return &apperrors.TaskNotDispatchedError{TaskID: taskID, Status: task.Status}
		}

		machine, lockErr := lockMachineByCode(tx, task.RecommendedMachine)
		if lockErr != nil {
			return lockErr
		}
		switch machine.Status {
		case constants.MachineRepair:
			return &apperrors.MachineUnderRepairError{MachineCode: machine.Code}
		case constants.MachineWorking:
			return fmt.Errorf("machine %s is already working", machine.Code)
		}

		task.Status = constants.TaskDispatched
		if err := tx.Save(task).Error; err != nil {
			return fmt.Errorf("mark task dispatched: %w", err)
		}
		machine.Status = constants.MachineWorking
		machine.CurrentTask = fmt.Sprintf("%s %s", task.Type, task.Field)
		if err := tx.Save(machine).Error; err != nil {
			return fmt.Errorf("mark machine working: %w", err)
		}
		if driver, findErr := findDriverByNameForUpdate(tx, task.RecommendedDriver); findErr == nil {
			if driver.Status == constants.DriverBusy {
				return fmt.Errorf("driver %s is already working", driver.Name)
			}
			driver.Status = constants.DriverBusy
			if err := tx.Save(driver).Error; err != nil {
				return fmt.Errorf("mark driver busy: %w", err)
			}
		} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}
		machineCode, driverName = machine.Code, task.RecommendedDriver
		return nil
	})
	return machineCode, driverName, err
}

// CompleteTask 完工登记事务：
//  1. 行锁锁定任务，仅「已派单」允许完工（重复/并发第二次起冲突失败）；
//  2. 生成与任务一一对应的可追溯作业记录（task_id 唯一索引兜底）；
//  3. 任务置为已完成；农机按实际工时扣减保养剩余工时，扣至零转维修中，否则恢复空闲；
//  4. 驾驶员恢复可派单并累计当月作业面积；同步刷新该农机的保养提醒剩余工时。
//
// 任一步骤失败整体回滚，记录、任务、农机、驾驶员状态均不改变。
func (r *TaskRepository) CompleteTask(plan *TaskCompletionPlan) (*TaskCompletionResult, error) {
	result := &TaskCompletionResult{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		task, err := lockTask(tx, plan.TaskID)
		if err != nil {
			return err
		}
		if task.Status == constants.TaskDone {
			return &apperrors.TaskAlreadyCompletedError{TaskID: task.ID}
		}
		if task.Status != constants.TaskDispatched {
			return &apperrors.TaskNotDispatchedError{TaskID: task.ID, Status: task.Status}
		}

		machine, err := lockMachineByCode(tx, task.RecommendedMachine)
		if err != nil {
			return err
		}

		taskID := task.ID
		record := &model.WorkRecord{
			ID:          buildRecordID(plan.Now),
			TaskID:      &taskID,
			MachineCode: machine.Code,
			DriverName:  task.RecommendedDriver,
			WorkDate:    plan.WorkDate,
			TaskType:    task.Type,
			ActualHours: plan.ActualHours,
			FuelLiters:  plan.FuelLiters,
			AreaMu:      plan.AreaMu,
			FuelCost:    plan.FuelCost,
			CreatedAt:   plan.Now,
		}
		if err := tx.Create(record).Error; err != nil {
			// task_id 唯一索引兜底：极端并发下后到的事务在此冲突，视同重复完工，整体回滚。
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return &apperrors.TaskAlreadyCompletedError{TaskID: task.ID}
			}
			return fmt.Errorf("create work record: %w", err)
		}

		machine.WorkHours += plan.ActualHours
		machine.MaintenanceRemainingHours = deductHours(machine.MaintenanceRemainingHours, plan.ActualHours)
		maintenanceDue := machine.MaintenanceRemainingHours <= 0
		if maintenanceDue {
			machine.Status = constants.MachineRepair
			machine.CurrentTask = "保养维修"
		} else {
			machine.Status = constants.MachineIdle
			machine.CurrentTask = "可派单"
		}
		if err := tx.Save(machine).Error; err != nil {
			return fmt.Errorf("update machine after completion: %w", err)
		}

		completedAt := plan.Now
		task.Status = constants.TaskDone
		task.CompletedAt = &completedAt
		if err := tx.Save(task).Error; err != nil {
			return fmt.Errorf("mark task done: %w", err)
		}

		driverStatus := ""
		if driver, findErr := findDriverByNameForUpdate(tx, task.RecommendedDriver); findErr == nil {
			driver.Status = constants.DriverAvailable
			driver.MonthAreaMu += plan.AreaMu
			if err := tx.Save(driver).Error; err != nil {
				return fmt.Errorf("restore driver: %w", err)
			}
			driverStatus = driver.Status
		} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}

		if err := refreshMaintenance(tx, machine.Code, plan.ActualHours, plan.WorkDate, maintenanceDue); err != nil {
			return err
		}

		result.Record = record
		result.TaskID = task.ID
		result.TaskStatus = task.Status
		result.MachineCode = machine.Code
		result.MachineStatus = machine.Status
		result.RemainingHours = machine.MaintenanceRemainingHours
		result.MaintenanceDue = maintenanceDue
		result.DriverName = task.RecommendedDriver
		result.DriverStatus = driverStatus
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// lockTask 行锁锁定任务。
func lockTask(tx *gorm.DB, taskID string) (*model.FarmTask, error) {
	var task model.FarmTask
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, "id = ?", taskID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock task: %w", err)
	}
	return &task, nil
}

// lockMachineByCode 行锁按编号锁定农机。
func lockMachineByCode(tx *gorm.DB, code string) (*model.Machine, error) {
	var machine model.Machine
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&machine, "code = ?", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock machine: %w", err)
	}
	return &machine, nil
}

// findDriverByNameForUpdate 行锁按姓名锁定驾驶员。
func findDriverByNameForUpdate(tx *gorm.DB, name string) (*model.Driver, error) {
	var driver model.Driver
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&driver, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// refreshMaintenance 同步扣减该农机全部保养提醒的剩余工时，扣至零转为紧急；
// 农机剩余工时已为零且没有任何提醒行时补一条紧急保养提醒。
func refreshMaintenance(tx *gorm.DB, machineCode string, hours float64, workDate string, maintenanceDue bool) error {
	var reminders []model.MaintenanceReminder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Find(&reminders, "machine_code = ?", machineCode).Error; err != nil {
		return fmt.Errorf("lock maintenance reminders: %w", err)
	}
	for i := range reminders {
		rem := &reminders[i]
		rem.RemainingHours = deductHours(rem.RemainingHours, hours)
		if rem.RemainingHours <= 0 {
			rem.Level = constants.MaintenanceLevelDanger
			rem.DueDate = workDate
			rem.LastServiceRecord = constants.MaintenanceZeroDueRecord
		}
		if err := tx.Save(rem).Error; err != nil {
			return fmt.Errorf("update maintenance reminder: %w", err)
		}
	}
	if maintenanceDue && len(reminders) == 0 {
		reminder := &model.MaintenanceReminder{
			ID:                "s-auto-" + uuid.NewString()[:8],
			MachineCode:       machineCode,
			Title:             constants.MaintenanceZeroDueTitle,
			DueDate:           workDate,
			RemainingHours:    0,
			Level:             constants.MaintenanceLevelDanger,
			LastServiceRecord: constants.MaintenanceZeroDueRecord,
			CreatedAt:         time.Now(),
		}
		if err := tx.Create(reminder).Error; err != nil {
			return fmt.Errorf("create maintenance reminder: %w", err)
		}
	}
	return nil
}

// deductHours 扣减工时并在零处截断，不允许出现负数。
func deductHours(remaining, hours float64) float64 {
	return math.Max(0, remaining-hours)
}

// buildRecordID 生成 32 位作业记录 ID（r + UUID 前 31 位十六进制）。
func buildRecordID(now time.Time) string {
	return constants.RecordIDPrefix + uuid.NewString()[:31]
}
