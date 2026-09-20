package constants

// 保养相关默认值。
const (
	// FuelPricePerLiter 柴油参考单价（元/升），用于由油耗折算油耗成本。
	FuelPricePerLiter = 7.4
	// MaintenanceRemainingZero 保养剩余时长归零阈值，达到即进入维修中。
	MaintenanceRemainingZero = 0.0
)

// 保养提醒级别。
const (
	MaintenanceLevelNormal  = "normal"
	MaintenanceLevelWarning = "warning"
	MaintenanceLevelDanger  = "danger"
)

// 农机恢复空闲时当前任务列展示文案。
const (
	MachineIdleTaskText = "可派单"
)
