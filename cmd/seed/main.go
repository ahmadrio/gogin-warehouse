package main

import (
	"fmt"
	"log"
	"warehouse-management/internal/auth"
	"warehouse-management/internal/config"
	"warehouse-management/internal/database"
	"warehouse-management/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultPassword = "password123"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	db.Logger = logger.Default.LogMode(logger.Warn)

	err = db.Transaction(func(tx *gorm.DB) error {
		perms, err := seedPermissions(tx)
		if err != nil {
			return err
		}
		roles, err := seedRoles(tx, perms)
		if err != nil {
			return err
		}
		warehouses, err := seedWarehouses(tx)
		if err != nil {
			return err
		}
		return seedUsers(tx, roles, warehouses)
	})
	if err != nil {
		log.Fatalf("seed gagal: %v", err)
	}
	log.Println("seed selesai")
}

func seedPermissions(tx *gorm.DB) (map[string]model.Permission, error) {
	ops := map[string]model.Operation{}
	for _, name := range auth.AllOperations {
		op := model.Operation{Name: name}
		if err := tx.Where("name = ?", name).FirstOrCreate(&op).Error; err != nil {
			return nil, err
		}
		ops[name] = op
	}

	result := map[string]model.Permission{}
	for _, resName := range auth.AllResources {
		res := model.Resource{Name: resName}
		if err := tx.Where("name = ?", resName).FirstOrCreate(&res).Error; err != nil {
			return nil, err
		}

		for opName, op := range ops {
			p := model.Permission{
				OperationID: op.ID,
				ResourceID:  res.ID,
			}
			if err := tx.Where("operation_id = ? AND resource_id = ?", op.ID, res.ID).FirstOrCreate(&p).Error; err != nil {
				return nil, err
			}
			result[auth.Perm(resName, opName)] = p
		}
	}

	return result, nil
}

func seedRoles(tx *gorm.DB, perms map[string]model.Permission) (map[string]model.Role, error) {
	all := make([]string, 0, len(perms))
	for key := range perms {
		all = append(all, key)
	}

	defs := map[string][]string{auth.RoleSuperadmin: all}
	for name, list := range auth.RoleDefinisions {
		defs[name] = list
	}

	roles := map[string]model.Role{}
	for name, permNames := range defs {
		role := model.Role{Name: name}
		if err := tx.Where("name = ?", name).FirstOrCreate(&role).Error; err != nil {
			return nil, err
		}

		for _, pn := range permNames {
			p, ok := perms[pn]
			if !ok {
				return nil, fmt.Errorf("permission %s not found", pn)
			}

			rp := model.RolePermission{
				RoleID:       role.ID,
				PermissionID: p.ID,
			}
			if err := tx.Where("role_id = ? AND permission_id = ?", role.ID, p.ID).FirstOrCreate(&rp).Error; err != nil {
				return nil, err
			}
		}

		roles[name] = role
	}

	return roles, nil
}

func seedWarehouses(tx *gorm.DB) (map[string]model.Warehouse, error) {
	data := []struct {
		Code, Name string
		Locations  []string
	}{
		{"WH-MALANG-01", "Gudang Malang Pakis", []string{"RAK-A-01", "RAK-A-02", "RAK-A-03"}},
		{"WH-SURABAYA-01", "Gudang Surabaya Rungkut", []string{"RAK-B-01", "RAK-B-02", "RAK-B-03"}},
		{"WH-SURABAYA-02", "Gudang Surabaya Sudirman", []string{"RAK-C-01", "RAK-C-02", "RAK-C-03"}},
	}

	result := map[string]model.Warehouse{}
	for _, d := range data {
		wh := model.Warehouse{
			Code:     d.Code,
			Name:     d.Name,
			IsActive: true,
		}
		if err := tx.Where("code = ?", d.Code).FirstOrCreate(&wh).Error; err != nil {
			return nil, err
		}

		for _, locCode := range d.Locations {
			loc := model.WarehouseLocation{
				WarehouseID: wh.ID,
				Code:        locCode,
				Name:        "Rak " + locCode,
				IsActive:    true,
			}
			if err := tx.Where("warehouse_id = ? AND code = ?", wh.ID, locCode).FirstOrCreate(&loc).Error; err != nil {
				return nil, err
			}
		}
		result[d.Code] = wh
	}

	return result, nil
}

func seedUsers(tx *gorm.DB, roles map[string]model.Role, warehouses map[string]model.Warehouse) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []struct {
		Email, Name, Role string
		Warehouses        []string
	}{
		{"superadmin@warehouse.test", "Super Admin", "superadmin", nil},
		{"supervisor@warehouse.test", "Budi Supervisor", "supervisor", []string{"WH-MALANG-01", "WH-SURABAYA-01"}},
		{"staff.malang@warehouse.test", "Sari Staff Malang", "staff", []string{"WH-MALANG-01"}},
		{"staff.surabaya@warehouse.test", "Andi Staff Surabaya", "staff", []string{"WH-SURABAYA-01"}},
	}

	for _, u := range users {
		user := model.User{Email: u.Email, Name: u.Name, Password: string(hash), IsActive: true}
		if err := tx.Where("email = ?", u.Email).FirstOrCreate(&user).Error; err != nil {
			return err
		}

		role := roles[u.Role]
		ur := model.UserRole{UserID: user.ID, RoleID: role.ID}
		if err := tx.Where("user_id = ? AND role_id = ?", user.ID, role.ID).FirstOrCreate(&ur).Error; err != nil {
			return err
		}

		for _, code := range u.Warehouses {
			wh := warehouses[code]
			uw := model.UserWarehouse{UserID: user.ID, WarehouseID: wh.ID}
			if err := tx.Where("user_id = ? AND warehouse_id = ?", user.ID, wh.ID).FirstOrCreate(&uw).Error; err != nil {
				return err
			}
		}
		log.Printf("user: %-32s role: %-10s password: %s", u.Email, u.Role, defaultPassword)
	}
	return nil
}
