package service

import (
    "context"
    "math"
    "time"

    "github.com/google/uuid"

    "backend/internal/dto"
    "backend/internal/model"
    "backend/internal/repository"
)

type ContactService interface {
    Create(ctx context.Context, input dto.ContactCreateRequest) (*dto.ContactResponse, error)
    GetByID(ctx context.Context, id string) (*dto.ContactResponse, error)
    Update(ctx context.Context, id string, input dto.ContactUpdateRequest) (*dto.ContactResponse, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, q dto.ContactQuery) (*dto.ContactListResponse, error)
    Merge(ctx context.Context, userID string, input dto.MergeRequest) (*dto.ContactResponse, error)
    GetGroupMembers(ctx context.Context, groupID string, page, pageSize int) (*dto.GroupMembersResponse, error)
    GetGroupHistory(ctx context.Context, groupID string, page, pageSize int) (*dto.GroupHistoryResponse, error)
    AddHistory(ctx context.Context, contactID, userID string, input dto.AddHistoryRequest) (*dto.ContactHistoryResponse, error)
    ListHistory(ctx context.Context, contactID string) ([]dto.ContactHistoryResponse, error)
    UpdateDnd(ctx context.Context, id string, input dto.DndUpdateRequest) (*dto.ContactResponse, error)
}

type contactService struct {
    repo    repository.ContactRepository
    tagRepo repository.TagRepository
    cfRepo  repository.CustomFieldRepository
}

func NewContactService(r repository.ContactRepository, tr repository.TagRepository, cfr repository.CustomFieldRepository) ContactService {
    return &contactService{repo: r, tagRepo: tr, cfRepo: cfr}
}

func toTagResponse(t model.Tag) dto.TagResponse {
    return dto.TagResponse{ID: t.ID, Name: t.Name, Color: t.Color}
}

func toCustomValueResponse(v model.CustomFieldValue, fieldName string) dto.CustomValueResponse {
    return dto.CustomValueResponse{FieldID: v.CustomFieldID, FieldName: fieldName, Value: v.Value}
}

func toContactResponse(c *model.Contact, fieldMap map[string]string) dto.ContactResponse {
    tags := make([]dto.TagResponse, 0, len(c.Tags))
    for _, t := range c.Tags {
        tags = append(tags, toTagResponse(t))
    }
    cvs := make([]dto.CustomValueResponse, 0, len(c.CustomValues))
    for _, v := range c.CustomValues {
        name := fieldMap[v.CustomFieldID]
        cvs = append(cvs, toCustomValueResponse(v, name))
    }
    return dto.ContactResponse{
        ID:           c.ID,
        Name:         c.Name,
        Email:        c.Email,
        Phone:        c.Phone,
        Company:      c.Company,
        Status:       c.Status,
        Notes:        c.Notes,
        IsGroup:      c.IsGroup,
        GroupID:      c.GroupID,
        LeadScore:    c.LeadScore,
        DndActive:    c.DndActive,
        DndType:      c.DndType,
        DndReason:    c.DndReason,
        DndUntil:     c.DndUntil,
        Tags:         tags,
        CustomValues: cvs,
        CreatedAt:    c.CreatedAt,
        UpdatedAt:    c.UpdatedAt,
    }
}

func toHistoryResponse(h *model.ContactHistory) dto.ContactHistoryResponse {
    return dto.ContactHistoryResponse{
        ID:          h.ID,
        ContactID:   h.ContactID,
        Action:      h.Action,
        Description: h.Description,
        UserID:      h.UserID,
        CreatedAt:   h.CreatedAt,
    }
}

func (s *contactService) fieldMap(ctx context.Context) map[string]string {
    fields, _ := s.cfRepo.List(ctx)
    m := make(map[string]string, len(fields))
    for _, f := range fields {
        m[f.ID] = f.Name
    }
    return m
}

func (s *contactService) refreshScore(ctx context.Context, c *model.Contact) {
    count, _ := s.repo.CountHistory(ctx, c.ID)
    lastActivity := -1
    last, _ := s.repo.LastHistoryTime(ctx, c.ID)
    if last != nil {
        lastActivity = daysSince(last.CreatedAt)
    }
    c.LeadScore = calcLeadScore(c, int(count), lastActivity)
}

func (s *contactService) applyTags(ctx context.Context, contactID string, tagIDs []string) error {
    if tagIDs == nil {
        return nil
    }
    tags, err := s.tagRepo.FindByIDs(ctx, tagIDs)
    if err != nil {
        return err
    }
    return s.repo.SetTags(ctx, contactID, tags)
}

func (s *contactService) applyCustomValues(ctx context.Context, contactID string, values map[string]string) error {
    if values == nil {
        return nil
    }
    cvs := make([]model.CustomFieldValue, 0, len(values))
    for fieldID, val := range values {
        cvs = append(cvs, model.CustomFieldValue{
            ID:            uuid.NewString(),
            ContactID:     contactID,
            CustomFieldID: fieldID,
            Value:         val,
        })
    }
    return s.repo.SetCustomValues(ctx, contactID, cvs)
}

func (s *contactService) Create(ctx context.Context, input dto.ContactCreateRequest) (*dto.ContactResponse, error) {
    debugLog("Contact.Create", "name", input.Name, "email", input.Email, "status", input.Status)
    status := input.Status
    if status == "" {
        status = "lead"
    }
    c := &model.Contact{
        ID:      uuid.NewString(),
        Name:    input.Name,
        Email:   input.Email,
        Phone:   input.Phone,
        Company: input.Company,
        Status:  status,
        Notes:   input.Notes,
    }
    userID, _ := ctx.Value("userID").(string)
    _ = s.repo.AddHistory(ctx, &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   c.ID,
        Action:      "created",
        Description: "Kontakt został utworzony",
        UserID:      userID,
        CreatedAt:   time.Now(),
    })
    s.refreshScore(ctx, c)
    if err := s.repo.Create(ctx, c); err != nil {
        return nil, ErrInternal
    }
    _ = s.applyTags(ctx, c.ID, input.TagIDs)
    _ = s.applyCustomValues(ctx, c.ID, input.CustomValues)
    updated, _ := s.repo.FindByID(ctx, c.ID)
    if updated == nil {
        updated = c
    }
    debugLogResult("Contact.Create", nil, "id", updated.ID)
    fm := s.fieldMap(ctx)
    r := toContactResponse(updated, fm)
    return &r, nil
}

func (s *contactService) GetByID(ctx context.Context, id string) (*dto.ContactResponse, error) {
    debugLog("Contact.GetByID", "id", id)
    c, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if c == nil {
        return nil, nil
    }
    debugLogResult("Contact.GetByID", nil, "found", c != nil)
    fm := s.fieldMap(ctx)
    r := toContactResponse(c, fm)
    return &r, nil
}

func (s *contactService) Update(ctx context.Context, id string, input dto.ContactUpdateRequest) (*dto.ContactResponse, error) {
    debugLog("Contact.Update", "id", id, "name", input.Name, "status", input.Status)
    c, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if c == nil {
        return nil, nil
    }
    oldStatus := c.Status
    if input.Name != "" {
        c.Name = input.Name
    }
    c.Email = input.Email
    c.Phone = input.Phone
    c.Company = input.Company
    if input.Status != "" {
        c.Status = input.Status
    }
    c.Notes = input.Notes

    userID, _ := ctx.Value("userID").(string)
    desc := "Kontakt został zaktualizowany"
    if input.Status != "" && input.Status != oldStatus {
        desc = "Status zmieniony z " + oldStatus + " na " + input.Status
    }
    _ = s.repo.AddHistory(ctx, &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   c.ID,
        Action:      "updated",
        Description: desc,
        UserID:      userID,
        CreatedAt:   time.Now(),
    })
    s.refreshScore(ctx, c)
    if err := s.repo.Update(ctx, c); err != nil {
        return nil, ErrInternal
    }
    _ = s.applyTags(ctx, c.ID, input.TagIDs)
    _ = s.applyCustomValues(ctx, c.ID, input.CustomValues)
    updated, _ := s.repo.FindByID(ctx, c.ID)
    if updated == nil {
        updated = c
    }
    debugLogResult("Contact.Update", nil, "id", id)
    fm := s.fieldMap(ctx)
    r := toContactResponse(updated, fm)
    return &r, nil
}

func (s *contactService) UpdateDnd(ctx context.Context, id string, input dto.DndUpdateRequest) (*dto.ContactResponse, error) {
    debugLog("Contact.UpdateDnd", "id", id, "active", input.DndActive, "type", input.DndType)
    c, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if c == nil {
        return nil, nil
    }
    c.DndActive = input.DndActive
    c.DndType = input.DndType
    c.DndReason = input.DndReason
    c.DndUntil = input.DndUntil
    if err := s.repo.Update(ctx, c); err != nil {
        return nil, ErrInternal
    }
    updated, _ := s.repo.FindByID(ctx, id)
    if updated == nil {
        updated = c
    }
    fm := s.fieldMap(ctx)
    r := toContactResponse(updated, fm)
    return &r, nil
}

func (s *contactService) Delete(ctx context.Context, id string) error {
    debugLog("Contact.Delete", "id", id)
    err := s.repo.Delete(ctx, id)
    debugLogResult("Contact.Delete", err)
    return err
}

func (s *contactService) List(ctx context.Context, q dto.ContactQuery) (*dto.ContactListResponse, error) {
    debugLog("Contact.List", "search", q.Search, "status", q.Status, "page", q.Page, "pageSize", q.PageSize)
    pageSize := q.PageSize
    if pageSize <= 0 {
        pageSize = 20
    }
    page := q.Page
    if page <= 0 {
        page = 1
    }

    contacts, total, err := s.repo.List(ctx, q.Search, q.Status, q.Tags, q.SortBy, q.SortDir, page, pageSize)
    if err != nil {
        debugLogResult("Contact.List", err)
        return nil, ErrInternal
    }
    debugLogResult("Contact.List", nil, "total", total, "returned", len(contacts))
    fm := s.fieldMap(ctx)
    items := make([]dto.ContactResponse, 0, len(contacts))
    for _, c := range contacts {
        items = append(items, toContactResponse(c, fm))
    }
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    return &dto.ContactListResponse{
        Data:       items,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

func (s *contactService) Merge(ctx context.Context, userID string, input dto.MergeRequest) (*dto.ContactResponse, error) {
    debugLog("Contact.Merge", "user_id", userID, "group_name", input.GroupName, "contact_count", len(input.ContactIDs))
    members, err := s.repo.FindByIDs(ctx, input.ContactIDs)
    if err != nil {
        return nil, ErrInternal
    }
    if len(members) < 2 {
        return nil, ErrInternal
    }

    status := input.Status
    if status == "" {
        status = "customer"
    }
    company := input.Company
    if company == "" && len(members) > 0 {
        company = members[0].Company
    }

    group := &model.Contact{
        ID:      uuid.NewString(),
        Name:    input.GroupName,
        Company: company,
        Status:  status,
        IsGroup: true,
    }

    for _, m := range members {
        m.GroupID = group.ID
        _ = s.repo.Update(ctx, m)
    }

    _ = s.repo.AddHistory(ctx, &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   group.ID,
        Action:      "merged",
        Description: "Połączono " + string(rune('0'+len(members))) + " kontaktów w grupę",
        UserID:      userID,
        CreatedAt:   time.Now(),
    })
    s.refreshScore(ctx, group)
    if err := s.repo.Create(ctx, group); err != nil {
        return nil, ErrInternal
    }

    debugLogResult("Contact.Merge", nil, "group_id", group.ID)
    fm := s.fieldMap(ctx)
    r := toContactResponse(group, fm)
    return &r, nil
}

func (s *contactService) GetGroupMembers(ctx context.Context, groupID string, page, pageSize int) (*dto.GroupMembersResponse, error) {
    debugLog("Contact.GetGroupMembers", "group_id", groupID, "page", page)
    if pageSize <= 0 {
        pageSize = 10
    }
    if page <= 0 {
        page = 1
    }
    members, total, err := s.repo.FindMembers(ctx, groupID, page, pageSize)
    if err != nil {
        return nil, ErrInternal
    }
    fm := s.fieldMap(ctx)
    items := make([]dto.ContactResponse, 0, len(members))
    for _, m := range members {
        items = append(items, toContactResponse(m, fm))
    }
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    return &dto.GroupMembersResponse{
        Data:       items,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

func (s *contactService) GetGroupHistory(ctx context.Context, groupID string, page, pageSize int) (*dto.GroupHistoryResponse, error) {
    debugLog("Contact.GetGroupHistory", "group_id", groupID, "page", page)
    if pageSize <= 0 {
        pageSize = 10
    }
    if page <= 0 {
        page = 1
    }
    items, total, err := s.repo.ListGroupHistory(ctx, groupID, page, pageSize)
    if err != nil {
        return nil, ErrInternal
    }
    result := make([]dto.ContactHistoryResponse, 0, len(items))
    for _, h := range items {
        result = append(result, toHistoryResponse(h))
    }
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    return &dto.GroupHistoryResponse{
        Data:       result,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

func (s *contactService) AddHistory(ctx context.Context, contactID, userID string, input dto.AddHistoryRequest) (*dto.ContactHistoryResponse, error) {
    debugLog("Contact.AddHistory", "contact_id", contactID, "action", input.Action)
    h := &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   contactID,
        Action:      input.Action,
        Description: input.Description,
        UserID:      userID,
    }
    if err := s.repo.AddHistory(ctx, h); err != nil {
        return nil, ErrInternal
    }
    c, err := s.repo.FindByID(ctx, contactID)
    if err == nil && c != nil {
        s.refreshScore(ctx, c)
        _ = s.repo.Update(ctx, c)
    }
    r := toHistoryResponse(h)
    return &r, nil
}

func (s *contactService) ListHistory(ctx context.Context, contactID string) ([]dto.ContactHistoryResponse, error) {
    debugLog("Contact.ListHistory", "contact_id", contactID)
    items, err := s.repo.ListHistory(ctx, contactID)
    if err != nil {
        return nil, ErrInternal
    }
    result := make([]dto.ContactHistoryResponse, 0, len(items))
    for _, h := range items {
        result = append(result, toHistoryResponse(h))
    }
    return result, nil
}
