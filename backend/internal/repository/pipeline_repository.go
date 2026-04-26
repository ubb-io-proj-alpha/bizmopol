package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "backend/internal/model"
)

type PipelineRepository interface {
    CreatePipeline(ctx context.Context, p *model.Pipeline) error
    FindPipelineByID(ctx context.Context, id string) (*model.Pipeline, error)
    UpdatePipeline(ctx context.Context, p *model.Pipeline) error
    DeletePipeline(ctx context.Context, id string) error
    ListPipelines(ctx context.Context) ([]*model.Pipeline, error)

    CreateStage(ctx context.Context, s *model.Stage) error
    FindStageByID(ctx context.Context, id string) (*model.Stage, error)
    UpdateStage(ctx context.Context, s *model.Stage) error
    DeleteStage(ctx context.Context, id string) error
    ListStages(ctx context.Context, pipelineID string) ([]*model.Stage, error)

    FindContactStage(ctx context.Context, contactID, pipelineID string) (*model.ContactStage, error)
    UpsertContactStage(ctx context.Context, cs *model.ContactStage) error
    ListContactsByStage(ctx context.Context, stageID string) ([]*model.ContactStage, error)
    ListContactStagesForPipeline(ctx context.Context, pipelineID string) ([]*model.ContactStage, error)
}

type pipelineRepository struct {
    db *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) PipelineRepository {
    return &pipelineRepository{db: db}
}

func (r *pipelineRepository) CreatePipeline(ctx context.Context, p *model.Pipeline) error {
    return r.db.WithContext(ctx).Create(p).Error
}

func (r *pipelineRepository) FindPipelineByID(ctx context.Context, id string) (*model.Pipeline, error) {
    var p model.Pipeline
    err := r.db.WithContext(ctx).Preload("Stages", func(db *gorm.DB) *gorm.DB {
        return db.Order("sort_order ASC")
    }).First(&p, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &p, nil
}

func (r *pipelineRepository) UpdatePipeline(ctx context.Context, p *model.Pipeline) error {
    return r.db.WithContext(ctx).Save(p).Error
}

func (r *pipelineRepository) DeletePipeline(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.Pipeline{}, "id = ?", id).Error
}

func (r *pipelineRepository) ListPipelines(ctx context.Context) ([]*model.Pipeline, error) {
    var pipelines []*model.Pipeline
    err := r.db.WithContext(ctx).Preload("Stages", func(db *gorm.DB) *gorm.DB {
        return db.Order("sort_order ASC")
    }).Order("created_at ASC").Find(&pipelines).Error
    return pipelines, err
}

func (r *pipelineRepository) CreateStage(ctx context.Context, s *model.Stage) error {
    return r.db.WithContext(ctx).Create(s).Error
}

func (r *pipelineRepository) FindStageByID(ctx context.Context, id string) (*model.Stage, error) {
    var s model.Stage
    err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &s, nil
}

func (r *pipelineRepository) UpdateStage(ctx context.Context, s *model.Stage) error {
    return r.db.WithContext(ctx).Save(s).Error
}

func (r *pipelineRepository) DeleteStage(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&model.Stage{}, "id = ?", id).Error
}

func (r *pipelineRepository) ListStages(ctx context.Context, pipelineID string) ([]*model.Stage, error) {
    var stages []*model.Stage
    err := r.db.WithContext(ctx).Where("pipeline_id = ?", pipelineID).Order("sort_order ASC").Find(&stages).Error
    return stages, err
}

func (r *pipelineRepository) FindContactStage(ctx context.Context, contactID, pipelineID string) (*model.ContactStage, error) {
    var cs model.ContactStage
    err := r.db.WithContext(ctx).Where("contact_id = ? AND pipeline_id = ?", contactID, pipelineID).First(&cs).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &cs, nil
}

func (r *pipelineRepository) UpsertContactStage(ctx context.Context, cs *model.ContactStage) error {
    existing, err := r.FindContactStage(ctx, cs.ContactID, cs.PipelineID)
    if err != nil {
        return err
    }
    if existing != nil {
        existing.StageID = cs.StageID
        return r.db.WithContext(ctx).Save(existing).Error
    }
    return r.db.WithContext(ctx).Create(cs).Error
}

func (r *pipelineRepository) ListContactsByStage(ctx context.Context, stageID string) ([]*model.ContactStage, error) {
    var items []*model.ContactStage
    err := r.db.WithContext(ctx).Where("stage_id = ?", stageID).Find(&items).Error
    return items, err
}

func (r *pipelineRepository) ListContactStagesForPipeline(ctx context.Context, pipelineID string) ([]*model.ContactStage, error) {
    var items []*model.ContactStage
    err := r.db.WithContext(ctx).Where("pipeline_id = ?", pipelineID).Find(&items).Error
    return items, err
}