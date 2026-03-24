package service

import (
	"backend/internal/model"
	"backend/internal/repository"
)

type PipelineService struct {
	repo *repository.PipelineRepository
}

func NewPipelineService(repo *repository.PipelineRepository) *PipelineService {
	return &PipelineService{repo: repo}
}

func (s *PipelineService) CreatePipeline(name string, userID uint) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{
		Name:   name,
		UserID: userID,
	}
	err := s.repo.CreatePipeline(pipeline)
	return pipeline, err
}

func (s *PipelineService) GetPipeline(id uint) (*model.Pipeline, error) {
	return s.repo.GetPipelineByID(id)
}

func (s *PipelineService) GetUserPipelines(userID uint) ([]model.Pipeline, error) {
	return s.repo.GetAllPipelines(userID)
}

func (s *PipelineService) UpdatePipeline(id uint, name string) error {
	pipeline := &model.Pipeline{ID: id, Name: name}
	return s.repo.UpdatePipeline(pipeline)
}

func (s *PipelineService) DeletePipeline(id uint) error {
	return s.repo.DeletePipeline(id)
}

// Stage services
func (s *PipelineService) CreateStage(name string, pipelineID uint, position int) (*model.Stage, error) {
	stage := &model.Stage{
		Name:       name,
		PipelineID: pipelineID,
		Position:   position,
	}
	err := s.repo.CreateStage(stage)
	return stage, err
}

func (s *PipelineService) UpdateStage(id uint, name string) error {
	stage := &model.Stage{ID: id, Name: name}
	return s.repo.UpdateStage(stage)
}

func (s *PipelineService) UpdateStagePosition(stageID uint, position int) error {
	return s.repo.UpdateStagePosition(stageID, position)
}

func (s *PipelineService) DeleteStage(id uint) error {
	return s.repo.DeleteStage(id)
}

// Lead services
func (s *PipelineService) CreateLead(name string, email string, phone string, stageID *uint) (*model.Lead, error) {
	lead := &model.Lead{
		Name:    name,
		Email:   email,
		Phone:   phone,
		StageID: stageID,
	}
	err := s.repo.CreateLead(lead)
	return lead, err
}

func (s *PipelineService) UpdateLead(id uint, name string, email string, phone string) error {
	lead := &model.Lead{
		ID:    id,
		Name:  name,
		Email: email,
		Phone: phone,
	}
	return s.repo.UpdateLead(lead)
}

func (s *PipelineService) MoveLeadToStage(leadID uint, stageID *uint, position int) error {
	return s.repo.MoveLeadToStage(leadID, stageID, position)
}

func (s *PipelineService) DeleteLead(id uint) error {
	return s.repo.DeleteLead(id)
}
