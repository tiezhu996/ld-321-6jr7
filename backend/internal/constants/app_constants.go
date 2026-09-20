package constants

// 应用常量。
const (
	AppName                 = "agridispatch"
	APIVersion              = "v1"
	PageSizeDefault         = 10
	PageSizeMax             = 100
	OverviewCacheKey        = "agridispatch:overview"
	OverviewCacheTTLSeconds = 30
	WSPushIntervalSeconds   = 5
)

// 角色
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// 农机状态
const (
	MachineIdle    = "空闲"
	MachineWorking = "作业中"
	MachineRepair  = "维修中"
)

// 任务状态
const (
	TaskPending    = "待派单"
	TaskDispatched = "已派单"
	TaskDone       = "已完成"
)

// 驾驶员状态
const (
	DriverAvailable = "可派单"
	DriverBusy      = "作业中"
	DriverResting   = "休息"
)

// 保养提醒等级
const (
	MaintenanceLevelNormal  = "normal"
	MaintenanceLevelWarning = "warning"
	MaintenanceLevelDanger  = "danger"
)

// 完工业务默认值
const (
	// DefaultMaintenanceIntervalHours 新机保养剩余工时默认值（每作业 100 小时保养一次）。
	DefaultMaintenanceIntervalHours = 100.0
	// DefaultFuelPriceYuan 柴油单价（元/升），用于按油耗折算油耗成本。
	DefaultFuelPriceYuan = 7.4
	// WorkDateLayout 作业日期格式。
	WorkDateLayout = "2006-01-02"
	// MaintenanceZeroDueTitle 剩余工时扣减至零时的保养提醒标题。
	MaintenanceZeroDueTitle = "保养工时已用尽，需立即保养"
	// MaintenanceZeroDueRecord 剩余工时扣减至零时的最近维修记录描述。
	MaintenanceZeroDueRecord = "工时扣减至零，系统自动转入维修中"
	// MaintenanceDueDateLayout 保养到期提醒日期格式。
	MaintenanceDueDateLayout = "2006-01-02"
	// RecordIDPrefix 完工生成作业记录的 ID 前缀。
	RecordIDPrefix = "r"
)

// 错误码
const (
	CodeOK              = 0
	CodeBadRequest      = 40000
	CodeUnauthorized    = 40100
	CodeForbidden       = 40300
	CodeNotFound        = 40400
	CodeConflict        = 40900
	CodeInternalError   = 50000
	CodeTooManyRequests = 42900
)
