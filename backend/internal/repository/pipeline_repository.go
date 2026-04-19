package repository

import (
	"backend/internal/model"
	"gorm.io/gorm"
)

type PipelineRepository struct {
	db *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) *PipelineRepository {
	return &PipelineRepository{db: db}
}

func preloadPipelineGraph(db *gorm.DB) *gorm.DB {
	return db.Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).Preload("Stages.Leads", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	})
}

func (r *PipelineRepository) CreatePipeline(pipeline *model.Pipeline) error {
	return r.db.Create(pipeline).Error
}

func (r *PipelineRepository) GetPipelineByIDForUser(id uint, userID string) (*model.Pipeline, error) {
	var pipeline model.Pipeline
	err := preloadPipelineGraph(r.db).
		Where("id = ? AND user_id = ?", id, userID).
		First(&pipeline).Error
	return &pipeline, err
}

func (r *PipelineRepository) GetAllPipelines(userID string) ([]model.Pipeline, error) {
	var pipelines []model.Pipeline
	err := preloadPipelineGraph(r.db).
		Where("user_id = ?", userID).
		Find(&pipelines).Error
	return pipelines, err
}

func (r *PipelineRepository) UpdatePipeline(id uint, userID string, name string) error {
	result := r.db.Model(&model.Pipeline{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("name", name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PipelineRepository) DeletePipeline(id uint, userID string) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Pipeline{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Stage methods
func (r *PipelineRepository) CreateStage(stage *model.Stage) error {
	return r.db.Create(stage).Error
}

func (r *PipelineRepository) GetStageByID(id uint) (*model.Stage, error) {
	var stage model.Stage
	err := r.db.Preload("Leads", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).First(&stage, id).Error
	return &stage, err
}

func (r *PipelineRepository) GetStageByIDForPipeline(stageID uint, pipelineID uint, userID string) (*model.Stage, error) {
	var stage model.Stage
	err := r.db.Model(&model.Stage{}).
		Joins("JOIN pipelines ON pipelines.id = stages.pipeline_id").
		Where("stages.id = ? AND stages.pipeline_id = ? AND pipelines.user_id = ?", stageID, pipelineID, userID).
		First(&stage).Error
	return &stage, err
}

func (r *PipelineRepository) UpdateStagePosition(stageID uint, position int) error {
	return r.db.Model(&model.Stage{}).Where("id = ?", stageID).Update("position", position).Error
}

func (r *PipelineRepository) UpdateStage(stage *model.Stage) error {
	return r.db.Model(stage).Updates(stage).Error
}

func (r *PipelineRepository) DeleteStage(id uint) error {
	return r.db.Delete(&model.Stage{}, id).Error
}

// Lead methods
func (r *PipelineRepository) CreateLead(lead *model.Lead) error {
	return r.db.Create(lead).Error
}

func (r *PipelineRepository) GetLeadByID(id uint) (*model.Lead, error) {
	var lead model.Lead
	err := r.db.First(&lead, id).Error
	return &lead, err
}

func (r *PipelineRepository) GetLeadByIDForPipeline(leadID uint, pipelineID uint, userID string) (*model.Lead, error) {
	var lead model.Lead
	err := r.db.Model(&model.Lead{}).
		Joins("JOIN stages ON stages.id = leads.stage_id").
		Joins("JOIN pipelines ON pipelines.id = stages.pipeline_id").
		Where("leads.id = ? AND stages.pipeline_id = ? AND pipelines.user_id = ?", leadID, pipelineID, userID).
		First(&lead).Error
	return &lead, err
}

func (r *PipelineRepository) GetNextLeadPosition(stageID uint) (int, error) {
	var nextPosition int
	err := r.db.Model(&model.Lead{}).
		Where("stage_id = ?", stageID).
		Select("COALESCE(MAX(position), -1) + 1").
		Scan(&nextPosition).Error
	return nextPosition, err
}

func (r *PipelineRepository) MoveLeadToStage(leadID uint, stageID *uint, position int) error {
	return r.db.Model(&model.Lead{}).Where("id = ?", leadID).Updates(map[string]interface{}{
		"stage_id": stageID,
		"position": position,
	}).Error
}

func (r *PipelineRepository) UpdateLead(lead *model.Lead) error {
	return r.db.Model(lead).Updates(lead).Error
}

func (r *PipelineRepository) UpdateLeadPosition(leadID uint, position int) error {
	return r.db.Model(&model.Lead{}).Where("id = ?", leadID).Update("position", position).Error
}

func (r *PipelineRepository) DeleteLead(id uint) error {
	return r.db.Delete(&model.Lead{}, id).Error
}
