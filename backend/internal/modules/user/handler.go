package user

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetMe(ctx context.Context, _ api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	u, err := h.svc.Me(ctx)
	if err != nil {
		return nil, err
	}
	return api.GetMe200JSONResponse(toAPI(u)), nil
}

func toAPI(u User) api.User {
	return api.User{
		Id:            u.ID,
		Email:         openapi_types.Email(u.Email),
		EmailVerified: u.EmailVerified,
		Name:          u.Name,
		AvatarUrl:     u.AvatarURL,
		IsSuperadmin:  u.IsSuperadmin,
		CreatedAt:     u.CreatedAt,
	}
}
