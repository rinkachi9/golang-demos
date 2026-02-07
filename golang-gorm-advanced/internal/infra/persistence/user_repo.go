package persistence

import (
	"context"

	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/domain/model"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/domain/scopes"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, activeOnly bool, domain string) ([]model.User, error) {
	query := r.db.WithContext(ctx).Model(&model.User{})
	if activeOnly {
		query = query.Scopes(scopes.ActiveUsers)
	}
	if domain != "" {
		query = query.Scopes(scopes.EmailDomain(domain))
	}

	var users []model.User
	if err := query.Order("created_at desc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
