package workspace

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

// Handler implements the workspace operations of the API. The workspace of
// the path is already validated by the workspace middleware and is in the
// request scope, so the handlers do not read it from the request.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListWorkspaces(ctx context.Context, _ api.ListWorkspacesRequestObject) (api.ListWorkspacesResponseObject, error) {
	list, err := h.svc.ListMine(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]api.Workspace, len(list))
	for i, w := range list {
		items[i] = workspaceToAPI(w)
	}
	return api.ListWorkspaces200JSONResponse{Items: items}, nil
}

func (h *Handler) CreateWorkspace(ctx context.Context, req api.CreateWorkspaceRequestObject) (api.CreateWorkspaceResponseObject, error) {
	if req.Body == nil {
		return nil, apperror.Validation(apperror.CodeInvalidRequest, "missing body")
	}
	w, err := h.svc.Create(ctx, req.Body.Name)
	if err != nil {
		return nil, err
	}
	return api.CreateWorkspace201JSONResponse(workspaceToAPI(w)), nil
}

func (h *Handler) GetWorkspace(ctx context.Context, _ api.GetWorkspaceRequestObject) (api.GetWorkspaceResponseObject, error) {
	w, err := h.svc.Current(ctx)
	if err != nil {
		return nil, err
	}
	return api.GetWorkspace200JSONResponse(workspaceToAPI(w)), nil
}

func (h *Handler) UpdateWorkspace(ctx context.Context, req api.UpdateWorkspaceRequestObject) (api.UpdateWorkspaceResponseObject, error) {
	if req.Body == nil {
		return nil, apperror.Validation(apperror.CodeInvalidRequest, "missing body")
	}
	w, err := h.svc.Rename(ctx, req.Body.Name)
	if err != nil {
		return nil, err
	}
	return api.UpdateWorkspace200JSONResponse(workspaceToAPI(w)), nil
}

func (h *Handler) ListMembers(ctx context.Context, _ api.ListMembersRequestObject) (api.ListMembersResponseObject, error) {
	members, err := h.svc.Members(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]api.Member, len(members))
	for i, m := range members {
		items[i] = memberToAPI(m)
	}
	return api.ListMembers200JSONResponse{Items: items}, nil
}

func (h *Handler) UpdateMember(ctx context.Context, req api.UpdateMemberRequestObject) (api.UpdateMemberResponseObject, error) {
	if req.Body == nil {
		return nil, apperror.Validation(apperror.CodeInvalidRequest, "missing body")
	}
	m, err := h.svc.ChangeRole(ctx, req.UserId, Role(req.Body.Role))
	if err != nil {
		return nil, err
	}
	return api.UpdateMember200JSONResponse(memberToAPI(m)), nil
}

func (h *Handler) RemoveMember(ctx context.Context, req api.RemoveMemberRequestObject) (api.RemoveMemberResponseObject, error) {
	if err := h.svc.RemoveMember(ctx, req.UserId); err != nil {
		return nil, err
	}
	return api.RemoveMember204Response{}, nil
}

func workspaceToAPI(w WithRole) api.Workspace {
	return api.Workspace{
		Id:        w.ID,
		Name:      w.Name,
		MyRole:    api.Role(w.Role),
		CreatedAt: w.CreatedAt,
	}
}

func memberToAPI(m Member) api.Member {
	return api.Member{
		UserId:    m.UserID,
		Email:     openapi_types.Email(m.Profile.Email),
		Name:      m.Profile.Name,
		AvatarUrl: m.Profile.AvatarURL,
		Role:      api.Role(m.Role),
		JoinedAt:  m.CreatedAt,
	}
}
