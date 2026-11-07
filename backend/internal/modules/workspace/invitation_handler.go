package workspace

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

func (h *Handler) ListInvitations(ctx context.Context, _ api.ListInvitationsRequestObject) (api.ListInvitationsResponseObject, error) {
	list, err := h.invitations.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]api.Invitation, len(list))
	for i, inv := range list {
		items[i] = invitationToAPI(inv)
	}
	return api.ListInvitations200JSONResponse{Items: items}, nil
}

func (h *Handler) CreateInvitation(ctx context.Context, req api.CreateInvitationRequestObject) (api.CreateInvitationResponseObject, error) {
	if req.Body == nil {
		return nil, apperror.Validation(apperror.CodeInvalidRequest, "missing body")
	}
	address := ""
	if req.Body.Email != nil {
		address = string(*req.Body.Email)
	}
	created, err := h.invitations.Create(ctx, InvitationKind(req.Body.Kind), address, Role(req.Body.Role))
	if err != nil {
		return nil, err
	}
	resp := api.CreateInvitation201JSONResponse{Invitation: invitationToAPI(created.Invitation)}
	if created.URL != "" {
		resp.Url = &created.URL
	}
	return resp, nil
}

func (h *Handler) RevokeInvitation(ctx context.Context, req api.RevokeInvitationRequestObject) (api.RevokeInvitationResponseObject, error) {
	if err := h.invitations.Revoke(ctx, req.InvitationId); err != nil {
		return nil, err
	}
	return api.RevokeInvitation204Response{}, nil
}

func (h *Handler) GetInvitation(ctx context.Context, req api.GetInvitationRequestObject) (api.GetInvitationResponseObject, error) {
	p, err := h.invitations.Preview(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	out := api.GetInvitation200JSONResponse{
		WorkspaceName: p.WorkspaceName,
		Kind:          api.InvitationKind(p.Kind),
		Role:          api.Role(p.Role),
		InvitedBy:     p.InvitedByName,
		ExpiresAt:     p.ExpiresAt,
		Status:        api.InvitationPreviewStatus(p.Status),
	}
	if p.Email != nil {
		e := openapi_types.Email(*p.Email)
		out.Email = &e
	}
	return out, nil
}

func (h *Handler) AcceptInvitation(ctx context.Context, req api.AcceptInvitationRequestObject) (api.AcceptInvitationResponseObject, error) {
	w, err := h.invitations.Accept(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return api.AcceptInvitation200JSONResponse(workspaceToAPI(w)), nil
}

func invitationToAPI(inv Invitation) api.Invitation {
	out := api.Invitation{
		Id:        inv.ID,
		Kind:      api.InvitationKind(inv.Kind),
		Role:      api.Role(inv.Role),
		ExpiresAt: inv.ExpiresAt,
		CreatedAt: inv.CreatedAt,
	}
	if inv.Email != nil {
		e := openapi_types.Email(*inv.Email)
		out.Email = &e
	}
	return out
}
