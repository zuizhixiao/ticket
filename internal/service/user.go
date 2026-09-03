// Package service 业务编排层。
package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ticket/internal/config"
	"ticket/internal/model"
	"ticket/internal/pkg/jwt"
	"ticket/internal/pkg/wechat"
	"ticket/internal/repository"
)

// 领域错误(handler 据此映射 HTTP 响应)
var (
	ErrNicknameTaken   = errors.New("昵称已被占用")
	ErrUserNotFound    = errors.New("用户不存在")
	ErrWrongCredential = errors.New("昵称或密码错误")
	ErrBanned          = errors.New("账号已被禁用")
	ErrNicknameLen     = errors.New("昵称需为 2-24 个字符")
	ErrPasswordLen     = errors.New("密码需为 6-64 位")
	ErrWechatService   = errors.New("微信服务暂不可用")
	ErrNotSubscribed   = errors.New("请先关注公众号后再试")
	ErrAdminProtected  = errors.New("管理员账号不可冻结")
	ErrSelfOperate     = errors.New("不能对自己执行该操作")
)

const (
	minNickRunes = 2
	maxNickRunes = 24
	minPassword  = 6
	maxPassword  = 64
)

// ---------- 密码 ----------

// HashPassword bcrypt 加密。
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword 兼容新 bcrypt 与旧 MD5;返回 legacy=true 表示命中旧 MD5 需升级。
func CheckPassword(hash, plain string) (ok bool, legacy bool) {
	if len(hash) == 32 {
		sum := md5.Sum([]byte(plain))
		return hex.EncodeToString(sum[:]) == hash, true
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil, false
}

func validateNickname(n string) error {
	n = strings.TrimSpace(n)
	nr := utf8.RuneCountInString(n)
	if nr < minNickRunes || nr > maxNickRunes {
		return ErrNicknameLen
	}
	return nil
}

func validatePassword(p string) error {
	if len(p) < minPassword || len(p) > maxPassword {
		return ErrPasswordLen
	}
	return nil
}

func issueToken(user *model.User) (string, error) {
	cfg := config.C.Jwt
	return jwt.Generate(jwt.Config{Secret: cfg.Secret, ExpireSeconds: cfg.ExpireSeconds},
		user.Id, user.Nickname, user.Avatar, user.Openid, user.Role)
}

// ---------- 注册 / 登录 / 找回 ----------

func Register(db *gorm.DB, nickname, password string) (*model.User, error) {
	if err := validateNickname(nickname); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if _, err := repository.GetUserByNickname(db, nickname); err == nil {
		return nil, ErrNicknameTaken
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{Nickname: strings.TrimSpace(nickname), Password: hash, Role: model.RoleUser, Status: model.StatusNormal}
	if err := repository.CreateUser(db, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login 登录成功返回用户与 token。
func Login(db *gorm.DB, nickname, password string) (*model.User, string, error) {
	u, err := repository.GetUserByNickname(db, strings.TrimSpace(nickname))
	if err != nil {
		return nil, "", ErrWrongCredential
	}
	if u.Status != model.StatusNormal {
		return nil, "", ErrBanned
	}
	ok, legacy := CheckPassword(u.Password, password)
	if !ok {
		return nil, "", ErrWrongCredential
	}
	if legacy { // 旧 MD5 命中,自动升级为 bcrypt
		if hash, err := HashPassword(password); err == nil {
			_ = repository.UpdateUserFields(db, u.Id, map[string]any{"password": hash})
		}
	}
	now := time.Now().Unix()
	_ = repository.UpdateUserFields(db, u.Id, map[string]any{"lastLoginTime": now})
	token, err := issueToken(u)
	return u, token, err
}

func ResetPassword(db *gorm.DB, nickname, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	u, err := repository.GetUserByNickname(db, strings.TrimSpace(nickname))
	if err != nil {
		return ErrUserNotFound
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return repository.UpdateUserFields(db, u.Id, map[string]any{"password": hash, "updateTime": now})
}

// UpdateProfile 更新昵称/头像。
func UpdateProfile(db *gorm.DB, userId int, nickname, avatar string) error {
	fields := map[string]any{}
	now := time.Now().Unix()
	if nickname != "" {
		if err := validateNickname(nickname); err != nil {
			return err
		}
		if other, err := repository.GetUserByNickname(db, strings.TrimSpace(nickname)); err == nil && other.Id != userId {
			return ErrNicknameTaken
		}
		fields["nickname"] = strings.TrimSpace(nickname)
	}
	if avatar != "" {
		fields["avatar"] = avatar
	}
	if len(fields) == 0 {
		return nil
	}
	fields["updateTime"] = now
	return repository.UpdateUserFields(db, userId, fields)
}

// BootstrapAdmins 将逗号分隔的昵称设为管理员(幂等,仅对已存在用户生效)。
func BootstrapAdmins(db *gorm.DB, csv string) error {
	for _, name := range strings.Split(csv, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		u, err := repository.GetUserByNickname(db, name)
		if err != nil {
			continue
		}
		if u.Role != model.RoleAdmin {
			if err := repository.UpdateUserFields(db, u.Id, map[string]any{"role": model.RoleAdmin}); err != nil {
				return err
			}
			fmt.Printf("[admin] bootstrap promoted %q\n", name)
		}
	}
	return nil
}

// ---------- 微信公众号验证码登录 ----------

// WeChatLogin 用公众号下发的 6 位验证码兑换本地登录态。
func WeChatLogin(db *gorm.DB, code string, wc *wechat.Client) (*model.User, string, bool, error) {
	openid, err := VerifyCode(code)
	if err != nil {
		return nil, "", false, err
	}
	accessToken, err := wc.GetAccessToken()
	if err != nil {
		return nil, "", false, ErrWechatService
	}
	info, err := wc.GetUserInfoByOpenID(accessToken, openid)
	if err != nil {
		return nil, "", false, ErrWechatService
	}
	if info.Subscribe == 0 {
		return nil, "", false, ErrNotSubscribed
	}

	nickname := strings.TrimSpace(info.Nickname)
	if nickname == "" {
		nickname = "微信用户" + openid[len(openid)-4:]
	}
	isNew := false
	u, err := repository.GetUserByOpenid(db, openid)
	if err != nil {
		if !repository.IsNotFound(err) {
			return nil, "", false, err
		}
		u = &model.User{
			Nickname: nickname,
			Avatar:   info.HeadImgURL,
			Openid:   openid,
			Role:     model.RoleUser,
			Status:   model.StatusNormal,
		}
		if err := repository.CreateUser(db, u); err != nil {
			return nil, "", false, err
		}
		isNew = true
	} else {
		if u.Status != model.StatusNormal {
			return nil, "", false, ErrBanned
		}
		now := time.Now().Unix()
		_ = repository.UpdateUserFields(db, u.Id, map[string]any{
			"nickname": nickname, "avatar": info.HeadImgURL, "lastLoginTime": now, "updateTime": now,
		})
	}
	token, err := issueToken(u)
	return u, token, isNew, err
}

// ---------- 管理员:用户管理 ----------

// AdminListUsers 分页用户列表(昵称模糊搜索)。
func AdminListUsers(db *gorm.DB, keyword string, page, size int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	list, err := repository.ListUsers(db, keyword, page, size)
	if err != nil {
		return nil, 0, err
	}
	total, err := repository.CountUsers(db, keyword)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminSetUserStatus 冻结(0)/解冻(1)普通用户账号。
func AdminSetUserStatus(db *gorm.DB, operatorID, targetID int, status int) error {
	target, err := repository.GetUserByID(db, targetID)
	if err != nil {
		if repository.IsNotFound(err) {
			return ErrUserNotFound
		}
		return err
	}
	if target.Role == model.RoleAdmin {
		return ErrAdminProtected
	}
	if targetID == operatorID {
		return ErrSelfOperate
	}
	now := time.Now().Unix()
	return repository.UpdateUserFields(db, targetID, map[string]any{
		"status": status, "updateTime": now,
	})
}

// AdminResetUserPassword 重置任意用户(不含管理员保护?允许重置普通用户)密码。
func AdminResetUserPassword(db *gorm.DB, targetID int, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	target, err := repository.GetUserByID(db, targetID)
	if err != nil {
		if repository.IsNotFound(err) {
			return ErrUserNotFound
		}
		return err
	}
	if target.Role == model.RoleAdmin {
		return ErrAdminProtected
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return repository.UpdateUserFields(db, targetID, map[string]any{
		"password": hash, "updateTime": now,
	})
}
