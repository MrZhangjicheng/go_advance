package application

import "dc/internal/biz"

// 应用序列化器
type ApplicationRes struct {
	AppId      string `json:"appId"`
	Name       string `json:"name"`
	Logo       string `json:"logo"`
	CreatedTs  int64  `json:"createdTs"`
	ModifiedTs int64  `json:"modifiedTs"`
}

// BuildApplication 序列化应用对象
func BuildApplication(item *biz.Application) ApplicationRes {
	data := ApplicationRes{
		AppId: item.AppId,
		Name:  item.Name,
		Logo:  item.Logo,
		// CreatedTs:  util.JSUnixTs(item.CreatedAt),
		// ModifiedTs: util.JSUnixTs(item.UpdatedAt),
	}
	return data
}
