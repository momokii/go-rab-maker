package models

type MasterLaborType struct {
	LaborTypeId      int     `json:"labor_type_id"`
	UserId           int     `json:"user_id"`
	RoleName         string  `json:"role_name"`
	Unit             string  `json:"unit"`
	DefaultDailyWage float64 `json:"default_daily_wage"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type MasterLaborTypeCreate struct {
	RoleName         string  `json:"role_name" validate:"required,min=1,max=100"`
	Unit             string  `json:"unit" validate:"required,min=1,max=20"`
	DefaultDailyWage float64 `json:"default_daily_wage" validate:"required,gte=0"`
	UserId           int     `json:"user_id"`
}
