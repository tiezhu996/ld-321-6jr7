package dto

// CompleteTaskRequest 任务完工登记请求。
// 三个字段均为必填且必须大于 0；任一缺失或非法时拒绝完工，
// 不允许改动作业记录、任务、农机或驾驶员任何状态。
type CompleteTaskRequest struct {
	ActualHours float64 `json:"actualHours" binding:"required,gt=0"` // 实际工时（小时）
	FuelLiters  float64 `json:"fuelLiters" binding:"required,gt=0"`  // 实际油耗（升）
	AreaMu      float64 `json:"areaMu" binding:"required,gt=0"`      // 实际作业面积（亩）
}

// CompleteTaskResult 任务完工结果。
type CompleteTaskResult struct {
	TaskID         string  `json:"taskId"`
	Status         string  `json:"status"`
	RecordID       string  `json:"recordId"`
	MachineCode    string  `json:"machineCode"`
	MachineStatus  string  `json:"machineStatus"`
	RemainingHours float64 `json:"remainingHours"`
	DriverName     string  `json:"driverName"`
	DriverStatus   string  `json:"driverStatus"`
	ActualHours    float64 `json:"actualHours"`
	FuelLiters     float64 `json:"fuelLiters"`
	AreaMu         float64 `json:"areaMu"`
	FuelCost       float64 `json:"fuelCost"`
	MaintenanceDue bool    `json:"maintenanceDue"`
	Message        string  `json:"message"`
}
