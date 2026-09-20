package repository

import (
	"errors"
	"fmt"

	"github.com/agridispatch/agridispatch/internal/constants"
	"github.com/agridispatch/agridispatch/internal/model"
	"gorm.io/gorm"
)

// ErrMachineBlocked 农机处于维修中，阻止派单。
var ErrMachineBlocked = errors.New("machine blocked for dispatch (under repair)")

// DispatchPlan 派单状态变更计划。
type DispatchPlan struct {
	TaskID             string
	MachineCode        string
	MachineCurrentTask string
	DriverName         string
}

// ApplyDispatch 在事务内完成派单：
// 维修中农机（保养剩余时长归零）不得继续派单，条件更新命中 0 行时
// 返回 ErrMachineBlocked；成功时农机置作业中、驾驶员置作业中、任务置已派单。
func (r *DashboardRepository) ApplyDispatch(plan *DispatchPlan) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Machine{}).
			Where("code = ? AND status <> ?", plan.MachineCode, constants.MachineRepair).
			Updates(map[string]interface{}{
				"status":       constants.MachineWorking,
				"current_task": plan.MachineCurrentTask,
			})
		if res.Error != nil {
			return fmt.Errorf("mark machine working: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return ErrMachineBlocked
		}

		if err := tx.Model(&model.Driver{}).
			Where("name = ?", plan.DriverName).
			Update("status", constants.DriverBusy).Error; err != nil {
			return fmt.Errorf("mark driver busy: %w", err)
		}

		res2 := tx.Model(&model.FarmTask{}).
			Where("id = ? AND status = ?", plan.TaskID, constants.TaskPending).
			Update("status", constants.TaskDispatched)
		if res2.Error != nil {
			return fmt.Errorf("mark task dispatched: %w", res2.Error)
		}
		if res2.RowsAffected == 0 {
			return ErrTaskNotDispatchable
		}
		return nil
	})
}

// ErrTaskNotDispatchable 任务已被派单/完工，不能再次派单。
var ErrTaskNotDispatchable = errors.New("task not in dispatchable status")
