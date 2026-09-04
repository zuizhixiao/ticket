package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"ticket/internal/model"
	"ticket/internal/repository"
	"ticket/internal/storage"
)

// 模板相关领域错误
var (
	ErrTemplateURLRequired = errors.New("模板图片地址不能为空")
	ErrTemplateNotFound    = errors.New("模板不存在")
)

// TemplateInput 模板新增/编辑入参。
type TemplateInput struct {
	Name       string `json:"name"`
	Url        string `json:"url"`
	TitleColor string `json:"titleColor"`
	TextColor  string `json:"textColor"`
	Status     int    `json:"status"`
}

var colorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func sanitizeColor(v, def string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	if !colorRe.MatchString(v) {
		return def
	}
	return v
}

// PublicTemplates 编辑器模板列表(系统上架)。
func PublicTemplates(db *gorm.DB) ([]model.Template, error) {
	return repository.ListPublicTemplates(db)
}

// AdminCreateTemplate 新增系统模板(默认置顶展示)。
func AdminCreateTemplate(db *gorm.DB, in TemplateInput) (*model.Template, error) {
	if strings.TrimSpace(in.Url) == "" {
		return nil, ErrTemplateURLRequired
	}
	minSort, err := repository.MinTemplateSort(db)
	if err != nil {
		return nil, err
	}
	t := &model.Template{
		UserId:     0, // 系统
		Name:       strings.TrimSpace(in.Name),
		Url:        strings.TrimSpace(in.Url),
		TitleColor: sanitizeColor(in.TitleColor, "#ffffff"),
		TextColor:  sanitizeColor(in.TextColor, "#ffffff"),
		Status:     model.StatusNormal,
		Sort:       minSort - 1,
	}
	if err := repository.CreateTemplate(db, t); err != nil {
		return nil, err
	}
	return t, nil
}

// AdminReorderTemplates 按给定 id 顺序批量写入 sort(0..n)。
func AdminReorderTemplates(db *gorm.DB, ids []int) error {
	for i, id := range ids {
		if err := repository.UpdateTemplateFields(db, id, map[string]any{"sort": i}); err != nil {
			return err
		}
	}
	return nil
}

// AdminUpdateTemplate 更新模板字段(空值不更新;status 可选 1/2)。
func AdminUpdateTemplate(db *gorm.DB, id int, in TemplateInput) (*model.Template, error) {
	cur, err := repository.GetTemplateByID(db, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	fields := map[string]any{}
	if in.Name != "" {
		fields["name"] = in.Name
	}
	if in.Url != "" {
		fields["url"] = in.Url
	}
	if in.TitleColor != "" {
		fields["titleColor"] = sanitizeColor(in.TitleColor, cur.TitleColor)
	}
	if in.TextColor != "" {
		fields["textColor"] = sanitizeColor(in.TextColor, cur.TextColor)
	}
	if in.Status == model.StatusNormal || in.Status == model.StatusDelete {
		fields["status"] = in.Status
	}
	if len(fields) == 0 {
		return cur, nil
	}
	if err := repository.UpdateTemplateFields(db, id, fields); err != nil {
		return nil, err
	}
	return repository.GetTemplateByID(db, id)
}

// AdminDeleteTemplate 硬删除模板记录(尽力删除其图片对象,失败不影响删除)。
func AdminDeleteTemplate(db *gorm.DB, id int) error {
	t, err := repository.GetTemplateByID(db, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return ErrTemplateNotFound
		}
		return err
	}
	if storage.Default != nil && t.Url != "" {
		_ = storage.Default.DeleteByURL(context.Background(), t.Url)
	}
	return repository.DeleteTemplate(db, id)
}
