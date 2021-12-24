package data

import (
	"context"
	"dc/internal/biz"
	"dc/internal/data/po"
)

type applicationRepo struct {
	data *Data
}

func NewApplicationRepo(data *Data) biz.ApplicationRepo {
	return &applicationRepo{
		data: data,
	}
}

// func (al *applicationRepo) ListApplication(ctx context.Context) ([]*biz.Application, error) {
// 	// ps,err :=
// }

func (al *applicationRepo) CreateApplication(ctx context.Context, app *biz.Application) error {
	a := &po.Application{
		AppId:  app.AppId,
		Type:   app.Type,
		UserId: app.UserId,
		Name:   app.Name,
		Logo:   app.Logo,
	}
	err := po.Save(al.data.db, a)
	return err
}
