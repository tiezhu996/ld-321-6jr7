package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerJSON = `{
  "swagger": "2.0",
  "info": {
    "title": "农机调度管理系统 API",
    "description": "农机资源管理、作业任务调度、实时轨迹监控、作业统计与维修保养提醒。",
    "version": "1.0.0"
  },
  "basePath": "/api/v1",
  "schemes": ["http", "ws"],
  "paths": {
    "/auth/login": { "post": { "summary": "登录", "tags": ["auth"] } },
    "/auth/me": { "get": { "summary": "当前用户", "tags": ["auth"] } },
    "/dashboard/overview": { "get": { "summary": "调度看板总览", "tags": ["dashboard"] } },
    "/dashboard/tasks/{id}/dispatch": { "post": { "summary": "一键派单", "tags": ["dashboard"] } },
    "/dashboard/reports/work/export": { "get": { "summary": "作业报表导出", "tags": ["dashboard"] } },
    "/tasks/{id}/dispatch": { "post": { "summary": "一键派单（推荐空闲农机与驾驶员）", "tags": ["task"] } },
    "/tasks/{id}/complete": {
      "post": {
        "summary": "完工登记：登记实际工时/油耗/作业面积，生成可追溯作业记录，任务结束并释放农机与驾驶员；保养剩余工时扣至零时农机转维修中",
        "tags": ["task"],
        "consumes": ["application/json"],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "type": "string" },
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "required": ["actualHours", "fuelLiters", "areaMu"],
              "properties": {
                "actualHours": { "type": "number", "minimum": 0.01, "description": "实际工时（必填，大于 0）" },
                "fuelLiters": { "type": "number", "minimum": 0, "description": "油耗升数（必填，允许为 0）" },
                "areaMu": { "type": "number", "minimum": 0.01, "description": "作业面积/亩（必填，大于 0）" }
              }
            }
          }
        ],
        "responses": {
          "200": { "description": "完工成功，返回作业记录、任务/农机/驾驶员最新状态与保养剩余工时" },
          "400": { "description": "字段缺失或数值不合法，不改变任何状态" },
          "404": { "description": "任务不存在" },
          "409": { "description": "任务状态不允许完工（非已派单）或重复/并发完工，只允许成功一次" }
        }
      }
    }
  }
}`

// SwaggerJSON 提供 swagger.json。
func (h *HealthHandler) SwaggerJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, swaggerJSON)
}
