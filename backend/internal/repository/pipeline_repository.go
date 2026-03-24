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

func (r *PipelineRepository) CreatePipeline(pipeline *model.Pipeline) error {
	return r.db.Create(pipeline).Error
}

func (r *PipelineRepository) GetPipelineByID(id uint) (*model.Pipeline, error) {
	var pipeline model.Pipeline
	err := r.db.Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).Preload("Stages.Leads", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).First(&pipeline, id).Error
	return &pipeline, err
}

func (r *PipelineRepository) GetAllPipelines(userID uint) ([]model.Pipeline, error) {
	var pipelines []model.Pipeline
	err := r.db.Where("user_id = ?", userID).Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).Preload("Stages.Leads", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).Find(&pipelines).Error
	return pipelines, err
}

func (r *PipelineRepository) UpdatePipeline(pipeline *model.Pipeline) error {
	return r.db.Model(pipeline).Updates(pipeline).Error
}

func (r *PipelineRepository) DeletePipeline(id uint) error {
	return r.db.Delete(&model.Pipeline{}, id).Error
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
