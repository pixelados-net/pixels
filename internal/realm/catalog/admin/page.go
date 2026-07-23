package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/niflaot/pixels/internal/permission"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

// CreatePage creates one catalog page.
func (service *Service) CreatePage(ctx context.Context, input PageInput) (catalogmodel.Page, error) {
	page := catalogmodel.Page{ParentID: input.ParentID, Name: strings.TrimSpace(input.Name), Layout: strings.TrimSpace(input.Layout),
		IconColor: input.IconColor, IconImage: input.IconImage, RequiredNode: input.RequiredNode, OrderNum: input.OrderNum,
		Visible: input.Visible, Enabled: input.Enabled, ClubOnly: input.ClubOnly}
	if err := service.validatePage(ctx, page); err != nil {
		return catalogmodel.Page{}, err
	}
	created, err := service.store.CreatePage(ctx, page)
	if err != nil {
		return catalogmodel.Page{}, err
	}
	if err := service.refresh(ctx); err != nil {
		return catalogmodel.Page{}, err
	}

	return created, nil
}

// UpdatePage applies a partial catalog page update.
func (service *Service) UpdatePage(ctx context.Context, id int64, patch PagePatch) (catalogmodel.Page, error) {
	if id <= 0 {
		return catalogmodel.Page{}, ErrInvalidPage
	}
	page, found, err := service.store.FindPageByID(ctx, id)
	if err != nil {
		return catalogmodel.Page{}, err
	}
	if !found {
		return catalogmodel.Page{}, ErrPageNotFound
	}
	applyPagePatch(&page, patch)
	if page.ParentID != nil && *page.ParentID == page.ID {
		return catalogmodel.Page{}, ErrInvalidPage
	}
	if err := service.validatePage(ctx, page); err != nil {
		return catalogmodel.Page{}, err
	}
	updated, found, err := service.store.UpdatePage(ctx, page)
	if err != nil {
		return catalogmodel.Page{}, err
	}
	if !found {
		return catalogmodel.Page{}, ErrConflict
	}
	if err := service.refresh(ctx); err != nil {
		return catalogmodel.Page{}, err
	}

	return updated, nil
}

// validatePage validates catalog page fields and its optional parent.
func (service *Service) validatePage(ctx context.Context, page catalogmodel.Page) error {
	if page.Name == "" || page.Layout == "" || page.IconColor < 0 || page.IconImage < 0 {
		return ErrInvalidPage
	}
	if page.RequiredNode != nil && (!page.RequiredNode.Concrete() || !permission.Registered(*page.RequiredNode)) {
		return ErrInvalidPage
	}
	if page.ParentID == nil {
		return nil
	}
	if *page.ParentID <= 0 {
		return ErrInvalidPage
	}
	_, found, err := service.store.FindPageByID(ctx, *page.ParentID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Join(ErrInvalidPage, ErrPageNotFound)
	}

	return nil
}

// applyPagePatch applies present page patch fields.
func applyPagePatch(page *catalogmodel.Page, patch PagePatch) {
	if patch.ParentID != nil {
		page.ParentID = *patch.ParentID
	}
	if patch.Name != nil {
		page.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Layout != nil {
		page.Layout = strings.TrimSpace(*patch.Layout)
	}
	if patch.IconColor != nil {
		page.IconColor = *patch.IconColor
	}
	if patch.IconImage != nil {
		page.IconImage = *patch.IconImage
	}
	if patch.RequiredNode != nil {
		page.RequiredNode = *patch.RequiredNode
	}
	if patch.OrderNum != nil {
		page.OrderNum = *patch.OrderNum
	}
	if patch.Visible != nil {
		page.Visible = *patch.Visible
	}
	if patch.Enabled != nil {
		page.Enabled = *patch.Enabled
	}
	if patch.ClubOnly != nil {
		page.ClubOnly = *patch.ClubOnly
	}
}
