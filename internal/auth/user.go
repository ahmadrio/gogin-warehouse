package auth

import (
	"slices"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID
	Email string
	Name  string
	Roles []string

	permissions map[string]struct{}

	AllWarehouses bool
	WarehouseIDs  []uuid.UUID
}

func NewUser(id uuid.UUID, name, email string, roles, permissions []string, warehouseIDs []uuid.UUID) *User {
	set := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		set[p] = struct{}{}
	}
	return &User{
		ID:            id,
		Name:          name,
		Email:         email,
		Roles:         roles,
		permissions:   set,
		AllWarehouses: slices.Contains(roles, RoleSuperadmin),
		WarehouseIDs:  warehouseIDs,
	}
}

func (u *User) Can(permission string) bool {
	_, ok := u.permissions[permission]
	return ok
}

func (u *User) Permissions() []string {
	list := make([]string, 0, len(u.permissions))
	for p := range u.permissions {
		list = append(list, p)
	}
	sort.Strings(list)
	return list
}

func (u *User) CanAccessWarehouse(id uuid.UUID) bool {
	return u.AllWarehouses || slices.Contains(u.WarehouseIDs, id)
}

// WarehouseScope adalah "filter gudang" yang diteruskan ke repository.
type WarehouseScope struct {
	All bool
	IDs []uuid.UUID
}

func (u *User) WarehouseScope() WarehouseScope {
	return WarehouseScope{All: u.AllWarehouses, IDs: u.WarehouseIDs}
}

// hanya di panggil jika sudah melewati middleware
const contextKey = "auth_user"

func SetUser(c *gin.Context, u *User) { c.Set(contextKey, u) }

func CurrentUser(c *gin.Context) *User {
	return c.MustGet(contextKey).(*User)
}
