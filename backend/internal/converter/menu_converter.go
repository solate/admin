package converter

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// ModelToMenuInfo 将数据库模型转换为菜单信息 DTO
func ModelToMenuInfo(menu *model.Menu) *dto.MenuInfo {
	if menu == nil {
		return nil
	}

	sort := int16(menu.Sort)
	resource := "menu:" + menu.MenuID
	action := "*"

	return &dto.MenuInfo{
		MenuID:      menu.MenuID,
		Name:        menu.Name,
		Type:        "MENU",
		ParentID:    stringPtr(menu.ParentID),
		Resource:    &resource,
		Action:      &action,
		Path:        stringPtr(menu.Path),
		Component:   stringPtr(menu.Component),
		Redirect:    stringPtr(menu.Redirect),
		Icon:        stringPtr(menu.Icon),
		Sort:        &sort,
		Status:      menu.Status,
		Description: stringPtr(menu.Description),
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
	}
}

// ModelListToMenuInfoList 批量将数据库模型转换为菜单信息 DTO
func ModelListToMenuInfoList(menus []*model.Menu) []*dto.MenuInfo {
	if len(menus) == 0 {
		return nil
	}

	result := make([]*dto.MenuInfo, len(menus))
	for i, menu := range menus {
		result[i] = ModelToMenuInfo(menu)
	}
	return result
}

// stringPtr 辅助函数：返回字符串指针
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// int16Ptr 辅助函数：返回 int16 指针
func int16Ptr(i int32) *int16 {
	v := int16(i)
	return &v
}
