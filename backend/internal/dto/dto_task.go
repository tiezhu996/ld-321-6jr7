package dto

// CompleteTaskRequest 完工登记请求。
//
// 使用指针类型区分「字段缺失」与「填报零值」：JSON 中未携带的字段为 nil，
// 由业务层判定为字段缺失并拒绝；携带的数值再做正数校验。
// 油耗允许为 0（电动农机或怠速作业），工时与作业面积必须大于 0。
type CompleteTaskRequest struct {
	ActualHours *float64 `json:"actualHours" validate:"required,gt=0"`
	FuelLiters  *float64 `json:"fuelLiters" validate:"required,gte=0"`
	AreaMu      *float64 `json:"areaMu" validate:"required,gt=0"`
}
