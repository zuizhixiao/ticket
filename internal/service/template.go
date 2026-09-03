package service

import (
	"errors"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"ticket/internal/model"
	"ticket/internal/repository"
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

// AdminCreateTemplate 新增系统模板。
func AdminCreateTemplate(db *gorm.DB, in TemplateInput) (*model.Template, error) {
	if strings.TrimSpace(in.Url) == "" {
		return nil, ErrTemplateURLRequired
	}
	t := &model.Template{
		UserId:     0, // 系统
		Name:       strings.TrimSpace(in.Name),
		Url:        strings.TrimSpace(in.Url),
		TitleColor: sanitizeColor(in.TitleColor, "#ffffff"),
		TextColor:  sanitizeColor(in.TextColor, "#ffffff"),
		Status:     model.StatusNormal,
	}
	if err := repository.CreateTemplate(db, t); err != nil {
		return nil, err
	}
	return t, nil
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

// AdminDeleteTemplate 软删模板。
func AdminDeleteTemplate(db *gorm.DB, id int) error {
	if _, err := repository.GetTemplateByID(db, id); err != nil {
		if repository.IsNotFound(err) {
			return ErrTemplateNotFound
		}
		return err
	}
	return repository.UpdateTemplateFields(db, id, map[string]any{"status": model.StatusDelete})
}
