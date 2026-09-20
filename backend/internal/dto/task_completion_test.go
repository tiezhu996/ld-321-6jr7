package dto

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCompleteTaskRequestBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name string
		body map[string]interface{}
		ok   bool
	}{
		{"all valid", map[string]interface{}{"actualHours": 8.5, "fuelLiters": 43, "areaMu": 120}, true},
		{"missing hours", map[string]interface{}{"fuelLiters": 43, "areaMu": 120}, false},
		{"missing fuel", map[string]interface{}{"actualHours": 8.5, "areaMu": 120}, false},
		{"missing area", map[string]interface{}{"actualHours": 8.5, "fuelLiters": 43}, false},
		{"zero hours", map[string]interface{}{"actualHours": 0, "fuelLiters": 43, "areaMu": 120}, false},
		{"negative fuel", map[string]interface{}{"actualHours": 8.5, "fuelLiters": -1, "areaMu": 120}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req CompleteTaskRequest
			c, _ := gin.CreateTestContext(nil)
			c.Request = mustJSONPost(t, tc.body)
			err := c.ShouldBindJSON(&req)
			if tc.ok && err != nil {
				t.Fatalf("expected bind success, got %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected bind failure for %s", tc.name)
			}
		})
	}
}
