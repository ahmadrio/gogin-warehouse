package model

import "github.com/google/uuid"

type Operation struct {
	Base
	Name string `json:"name"`
}

type Resource struct {
	Base
	Name string `json:"name"`
}

type Permission struct {
	Base
	OperationID uuid.UUID `json:"operation_id"`
	ResourceID  uuid.UUID `json:"resource_id"`

	Operation Operation `json:"operation"`
	Resource  Resource  `json:"resource"`
}

type Role struct {
	Base
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type RolePermission struct {
	Base
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`

	Role       Role       `json:"role"`
	Permission Permission `json:"permission"`
}

type User struct {
	Base
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
	IsActive bool   `json:"is_active"`
}

type UserRole struct {
	Base
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`

	User User `json:"user"`
	Role Role `json:"role"`
}

type UserWarehouse struct {
	Base
	UserID      uuid.UUID `json:"user_id"`
	WarehouseID uuid.UUID `json:"warehouse_id"`

	User      User      `json:"user"`
	Warehouse Warehouse `json:"warehouse"`
}
