package po

import "gorm.io/gorm"

type Application struct {
	gorm.Model
	AppId   string `gorm:"type:varchar(128);not null"` // 应用ID
	Type    string `gorm:"type:varchar(32);not null"`  // 应用类型
	UserId  string `gorm:"type:varchar(255);not null"` // 创建者ID
	Name    string `gorm:"type:varchar(256);not null"` // 应用名称
	Logo    string `gorm:"type:varchar(255);not null"` // 应用logo
	Status  string `gorm:"type:varchar(32);not null"`  // 应用状态
	GroupID uint   `gorm:"type:int(10)"`               // 群组id
}

// 保存
func Save(db *gorm.DB, item *Application) error {
	err := db.Save(item).Error
	return err
}

// 删除
func Delete(db *gorm.DB, hard bool, item *Application) error {
	socpe := db
	if hard {
		socpe = socpe.Unscoped()
	}
	err := socpe.Delete(item).Error
	return err
}

// 读取应用完整信息
func QryApplication(db *gorm.DB, id string) (*Application, error) {
	var item Application
	result := db.Where(Application{AppId: id}).First(&item)
	return &item, result.Error
}
