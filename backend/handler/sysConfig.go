package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/kmoretti/moments/db"
	"github.com/kmoretti/moments/pkg/util"
	"github.com/kmoretti/moments/vo"
	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
	"gorm.io/gorm"
)

type SysConfigHandler struct {
	base BaseHandler
}

func NewSysConfigHandler(injector do.Injector) *SysConfigHandler {
	return &SysConfigHandler{do.MustInvoke[BaseHandler](injector)}
}

// GetConfig godoc
//
//	@Tags			SysConfig
//	@Summary		获取系统设置(部分不敏感的)
//	@Description	敏感信息不返回,包括各种key密钥
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	vo.SysConfigVO
//	@Router			/sysConfig/get [post]
func (s SysConfigHandler) GetConfig(c echo.Context) error {
	var (
		config db.SysConfig
		result vo.SysConfigVO
	)

	if err := s.base.db.First(&config).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return SuccessResp(c, h{})
	}
	err := json.Unmarshal([]byte(config.Content), &result)
	if err != nil {
		return FailRespWithMsg(c, Fail, "读取系统配置异常")
	}
	result.Version = s.base.cfg.Version
	result.CommitId = s.base.cfg.CommitId

	suffix := result.S3.ThumbnailSuffix
	result.S3 = vo.S3VO{
		ThumbnailSuffix: suffix,
	}
	return SuccessResp(c, result)
}

// GetFullConfig godoc
//
//	@Tags		SysConfig
//	@Summary	获取系统设置(完整的)
//	@Accept		json
//	@Produce	json
//	@Param		x-api-token	header		string	true	"登录TOKEN"
//	@Success	200			{object}	vo.FullSysConfigVO
//	@Success	200
//	@Router		/api/sysConfig/get [post]
func (s SysConfigHandler) GetFullConfig(c echo.Context) error {
	var (
		config db.SysConfig
		result vo.FullSysConfigVO
	)

	context := c.(CustomContext)
	currentUser := context.CurrentUser()
	if currentUser == nil || currentUser.Id != 1 {
		return FailRespWithMsg(c, Fail, "需要先登录")
	}
	if err := s.base.db.First(&config).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return SuccessResp(c, h{})
	}
	err := json.Unmarshal([]byte(config.Content), &result)
	if err != nil {
		return FailRespWithMsg(c, Fail, "读取系统配置异常")
	}
	result.Version = s.base.cfg.Version
	result.CommitId = s.base.cfg.CommitId
	return SuccessResp(c, result)
}

// SaveConfig godoc
//
//	@Tags		SysConfig
//	@Summary	保存系统设置
//	@Accept		json
//	@Produce	json
//	@Param		object		body	vo.FullSysConfigVO	true	"保存系统设置"
//	@Param		x-api-token	header	string				true	"登录TOKEN"
//	@Success	200
//	@Router		/api/sysConfig/save [post]
func validateMusicConfig(music vo.MusicItemVO) error {
	music.Url = strings.TrimSpace(music.Url)
	music.Cover = strings.TrimSpace(music.Cover)

	if music.Url == "" {
		return nil
	}

	if music.External {
		valid, _ := util.ValidHttpUrl(music.Url)
		if !valid {
			return errors.New("音乐外链格式不正确")
		}
	} else if !strings.HasPrefix(music.Url, "/upload/") {
		return errors.New("本地音乐必须先上传到服务器")
	}

	if music.Cover != "" && !strings.HasPrefix(music.Cover, "/upload/") {
		valid, _ := util.ValidHttpUrl(music.Cover)
		if !valid {
			return errors.New("封面地址格式不正确")
		}
	}

	return nil
}

func (s SysConfigHandler) SaveConfig(c echo.Context) error {
	var (
		config db.SysConfig
		result vo.FullSysConfigVO
	)
	context := c.(CustomContext)
	currentUser := context.CurrentUser()
	if currentUser == nil || currentUser.Id != 1 {
		return FailRespWithMsg(c, Fail, "需要先登录")
	}

	if err := c.Bind(&result); err != nil {
		s.base.log.Info().Msgf("保存配置错误,%s", err)
		return FailResp(c, ParamError)
	}

	if err := validateMusicConfig(result.Music); err != nil {
		return FailRespWithMsg(c, ParamError, err.Error())
	}

	data, err := json.Marshal(result)
	if err != nil {
		return FailRespWithMsg(c, Fail, "读取系统配置异常")
	}

	if err := s.base.db.First(&config).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		config.Content = string(data)
		if err = s.base.db.Save(&config).Error; err != nil {
			return FailRespWithMsg(c, Fail, "保存系统配置异常")
		}
	} else {
		config.Content = string(data)
		if err = s.base.db.Updates(&config).Error; err != nil {
			return FailRespWithMsg(c, Fail, "保存系统配置异常")
		}
	}
	s.base.db.Table("User").Where("id=?", 1).Update("username", result.AdminUserName)
	return SuccessResp(c, h{})
}
