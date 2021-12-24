package service

import (
	conf "dc/configs"
	"dc/internal/biz"
	"dc/internal/data"
	appSerializer "dc/internal/serializer/application"
	"dc/internal/serializer/common"

	"github.com/gin-gonic/gin"
)

type AppCreatRequest struct {
	Name       string `json:"name"  form:"name"`
	Logo       string `json:"logo" form:"logo"`
	TemplateId string `json:"templateId" form:"templateId"`
	GroupId    uint   `json:"groupId" form:"groupId"`
}

type DcApplicationService struct {
	application *biz.ApplicationUsecase `json:"-"`
	// 需要请求参数
	AppCreatRequest AppCreatRequest `json:"appCreatRequest"`
}

func (s *DcApplicationService) Add() {
	repo := data.NewApplicationRepo(conf.SourceDB)
	s.application = biz.NewApplicationUsecase(repo)
}

func (s *DcApplicationService) post(ctx *gin.Context) common.Response {
	// DTO 对象对应的入参和出参
	// DTO入参 转DO
	app := &biz.Application{
		Name: s.AppCreatRequest.Name,
		Logo: s.AppCreatRequest.Logo,
	}
	err := s.application.Create(ctx, app)
	if err != nil {
		return common.DBOperateErr(err)
	}
	// 将DO转成DTO出参
	resData := appSerializer.BuildApplication(app)
	return common.Response{
		Data: resData,
	}
}

func (s *DcApplicationService) Post(c *gin.Context) common.Response {
	s.Add()
	res := s.post(c)
	if res.Code != 0 {
		// 记录日志
	}
	return res
}
