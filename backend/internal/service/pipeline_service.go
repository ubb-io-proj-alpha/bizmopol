package service

import (
    "context"
    "time"

    "github.com/google/uuid"

    "backend/internal/dto"
    "backend/internal/model"
    "backend/internal/repository"
)

type PipelineService interface {
    CreatePipeline(ctx context.Context, input dto.PipelineCreateRequest) (*dto.PipelineResponse, error)
    GetPipeline(ctx context.Context, id string) (*dto.PipelineResponse, error)
    UpdatePipeline(ctx context.Context, id string, input dto.PipelineUpdateRequest) (*dto.PipelineResponse, error)
    DeletePipeline(ctx context.Context, id string) error
    ListPipelines(ctx context.Context) ([]*dto.PipelineResponse, error)

    CreateStage(ctx context.Context, pipelineID string, input dto.StageCreateRequest) (*dto.StageResponse, error)
    UpdateStage(ctx context.Context, id string, input dto.StageUpdateRequest) (*dto.StageResponse, error)
    DeleteStage(ctx context.Context, id string) error

    MoveContact(ctx context.Context, pipelineID string, input dto.MoveContactRequest, userID string) error
    GetKanban(ctx context.Context, pipelineID string) (*dto.KanbanResponse, error)
}

type pipelineService struct {
    repo        repository.PipelineRepository
    contactRepo repository.ContactRepository
}

func NewPipelineService(r repository.PipelineRepository, cr repository.ContactRepository) PipelineService {
    return &pipelineService{repo: r, contactRepo: cr}
}

func toStageResponse(s *model.Stage) dto.StageResponse {
    return dto.StageResponse{
        ID:         s.ID,
        PipelineID: s.PipelineID,
        Name:       s.Name,
        Color:      s.Color,
        SortOrder:  s.SortOrder,
        CreatedAt:  s.CreatedAt,
        UpdatedAt:  s.UpdatedAt,
    }
}

func toPipelineResponse(p *model.Pipeline) *dto.PipelineResponse {
    stages := make([]dto.StageResponse, 0, len(p.Stages))
    for i := range p.Stages {
        stages = append(stages, toStageResponse(&p.Stages[i]))
    }
    return &dto.PipelineResponse{
        ID:          p.ID,
        Name:        p.Name,
        Description: p.Description,
        Stages:      stages,
        CreatedAt:   p.CreatedAt,
        UpdatedAt:   p.UpdatedAt,
    }
}

func (s *pipelineService) CreatePipeline(ctx context.Context, input dto.PipelineCreateRequest) (*dto.PipelineResponse, error) {
    p := &model.Pipeline{
        ID:          uuid.NewString(),
        Name:        input.Name,
        Description: input.Description,
    }
    stages := make([]model.Stage, 0, len(input.Stages))
    for i, si := range input.Stages {
        color := si.Color
        if color == "" {
            color = "#38bdf8"
        }
        stages = append(stages, model.Stage{
            ID:         uuid.NewString(),
            PipelineID: p.ID,
            Name:       si.Name,
            Color:      color,
            SortOrder:  si.SortOrder,
        })
        if si.SortOrder == 0 {
            stages[i].SortOrder = i
        }
    }
    p.Stages = stages
    if err := s.repo.CreatePipeline(ctx, p); err != nil {
        return nil, ErrInternal
    }
    return toPipelineResponse(p), nil
}

func (s *pipelineService) GetPipeline(ctx context.Context, id string) (*dto.PipelineResponse, error) {
    p, err := s.repo.FindPipelineByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if p == nil {
        return nil, nil
    }
    return toPipelineResponse(p), nil
}

func (s *pipelineService) UpdatePipeline(ctx context.Context, id string, input dto.PipelineUpdateRequest) (*dto.PipelineResponse, error) {
    p, err := s.repo.FindPipelineByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if p == nil {
        return nil, nil
    }
    if input.Name != "" {
        p.Name = input.Name
    }
    p.Description = input.Description
    if err := s.repo.UpdatePipeline(ctx, p); err != nil {
        return nil, ErrInternal
    }
    updated, _ := s.repo.FindPipelineByID(ctx, id)
    if updated == nil {
        updated = p
    }
    return toPipelineResponse(updated), nil
}

func (s *pipelineService) DeletePipeline(ctx context.Context, id string) error {
    return s.repo.DeletePipeline(ctx, id)
}

func (s *pipelineService) ListPipelines(ctx context.Context) ([]*dto.PipelineResponse, error) {
    pipelines, err := s.repo.ListPipelines(ctx)
    if err != nil {
        return nil, ErrInternal
    }
    result := make([]*dto.PipelineResponse, 0, len(pipelines))
    for _, p := range pipelines {
        result = append(result, toPipelineResponse(p))
    }
    return result, nil
}

func (s *pipelineService) CreateStage(ctx context.Context, pipelineID string, input dto.StageCreateRequest) (*dto.StageResponse, error) {
    color := input.Color
    if color == "" {
        color = "#38bdf8"
    }
    st := &model.Stage{
        ID:         uuid.NewString(),
        PipelineID: pipelineID,
        Name:       input.Name,
        Color:      color,
        SortOrder:  input.SortOrder,
    }
    if err := s.repo.CreateStage(ctx, st); err != nil {
        return nil, ErrInternal
    }
    r := toStageResponse(st)
    return &r, nil
}

func (s *pipelineService) UpdateStage(ctx context.Context, id string, input dto.StageUpdateRequest) (*dto.StageResponse, error) {
    st, err := s.repo.FindStageByID(ctx, id)
    if err != nil {
        return nil, ErrInternal
    }
    if st == nil {
        return nil, nil
    }
    if input.Name != "" {
        st.Name = input.Name
    }
    if input.Color != "" {
        st.Color = input.Color
    }
    st.SortOrder = input.SortOrder
    if err := s.repo.UpdateStage(ctx, st); err != nil {
        return nil, ErrInternal
    }
    r := toStageResponse(st)
    return &r, nil
}

func (s *pipelineService) DeleteStage(ctx context.Context, id string) error {
    return s.repo.DeleteStage(ctx, id)
}

func (s *pipelineService) MoveContact(ctx context.Context, pipelineID string, input dto.MoveContactRequest, userID string) error {
    stage, err := s.repo.FindStageByID(ctx, input.StageID)
    if err != nil || stage == nil {
        return ErrInternal
    }

    existing, _ := s.repo.FindContactStage(ctx, input.ContactID, pipelineID)
    oldStageName := ""
    if existing != nil {
        oldStage, _ := s.repo.FindStageByID(ctx, existing.StageID)
        if oldStage != nil {
            oldStageName = oldStage.Name
        }
    }

    cs := &model.ContactStage{
        ID:         uuid.NewString(),
        ContactID:  input.ContactID,
        PipelineID: pipelineID,
        StageID:    input.StageID,
    }
    if err := s.repo.UpsertContactStage(ctx, cs); err != nil {
        return ErrInternal
    }

    desc := "Przeniesiono do etapu: " + stage.Name
    if oldStageName != "" {
        desc = "Przeniesiono z etapu \"" + oldStageName + "\" do \"" + stage.Name + "\""
    }
    _ = s.contactRepo.AddHistory(ctx, &model.ContactHistory{
        ID:          uuid.NewString(),
        ContactID:   input.ContactID,
        Action:      "stage_change",
        Description: desc,
        UserID:      userID,
        CreatedAt:   time.Now(),
    })
    return nil
}

func (s *pipelineService) GetKanban(ctx context.Context, pipelineID string) (*dto.KanbanResponse, error) {
    p, err := s.repo.FindPipelineByID(ctx, pipelineID)
    if err != nil || p == nil {
        return nil, ErrInternal
    }

    contactStages, err := s.repo.ListContactStagesForPipeline(ctx, pipelineID)
    if err != nil {
        return nil, ErrInternal
    }

    stageContactMap := make(map[string][]string)
    for _, cs := range contactStages {
        stageContactMap[cs.StageID] = append(stageContactMap[cs.StageID], cs.ContactID)
    }

    columns := make([]dto.KanbanColumnResponse, 0, len(p.Stages))
    for i := range p.Stages {
        stage := &p.Stages[i]
        contactIDs := stageContactMap[stage.ID]
        contacts := make([]dto.KanbanContactResponse, 0)
        if len(contactIDs) > 0 {
            found, err := s.contactRepo.FindByIDs(ctx, contactIDs)
            if err == nil {
                for _, c := range found {
                    contacts = append(contacts, dto.KanbanContactResponse{
                        ID:        c.ID,
                        Name:      c.Name,
                        Email:     c.Email,
                        Phone:     c.Phone,
                        Company:   c.Company,
                        Status:    c.Status,
                        LeadScore: c.LeadScore,
                        StageID:   stage.ID,
                        CreatedAt: c.CreatedAt,
                    })
                }
            }
        }
        columns = append(columns, dto.KanbanColumnResponse{
            Stage:    toStageResponse(stage),
            Contacts: contacts,
        })
    }

    return &dto.KanbanResponse{
        Pipeline: *toPipelineResponse(p),
        Columns:  columns,
    }, nil
}