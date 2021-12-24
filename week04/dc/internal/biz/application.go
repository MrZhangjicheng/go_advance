package biz

import (
	"context"
	"time"
)

type Application struct {
	AppId     string
	Type      string
	UserId    string
	Name      string
	Logo      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ApplicationRepo interface {
	// db
	// ListApplication(ctx context.Context) ([]*Application, error)
	// GetApplication(ctx context.Context, id string) (*Application, error)
	CreateApplication(ctx context.Context, application *Application) error
	// UpdateApplication(ctx context.Context, id string, application *Application) error
	// DeleteApplication(ctx context.Context, id string) error

	// redis
	// GetArticleLike(ctx context.Context, id int64) (rv int64, err error)
	// IncArticleLike(ctx context.Context, id int64) error
}

type ApplicationUsecase struct {
	repo ApplicationRepo
}

func NewApplicationUsecase(repo ApplicationRepo) *ApplicationUsecase {
	return &ApplicationUsecase{repo: repo}
}

// func (uc *ApplicationUsecase) List(ctx context.Context, appid string) ([]*Application, error) {
// 	ps, err := uc.repo.ListApplication(ctx)
// 	if err != nil {
// 		return []*Application{}, err
// 	}
// 	return ps, nil
// }

// func (uc *ApplicationUsecase) Get(ctx context.Context, appid string) (*Application, error) {
// 	p, err := uc.repo.GetApplication(ctx, appid)
// 	if err != nil {
// 		return &Application{}, err
// 	}
// 	return p, nil
// }

func (uc *ApplicationUsecase) Create(ctx context.Context, app *Application) error {
	return uc.repo.CreateApplication(ctx, app)
}

// func (uc *ApplicationUsecase) Update(ctx context.Context, appid string, app *Application) error {
// 	return uc.repo.UpdateApplication(ctx, appid, app)
// }

// func (uc *ApplicationUsecase) Delete(ctx context.Context, appid string) error {
// 	return uc.repo.DeleteApplication(ctx, appid)
// }
