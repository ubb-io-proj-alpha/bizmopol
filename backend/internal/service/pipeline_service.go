package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"errors"
)

type PipelineService struct {
	repo *repository.PipelineRepository
}

var ErrStageIDRequired = errors.New("stage_id_required")

func NewPipelineService(repo *repository.PipelineRepository) *PipelineService {
	return &PipelineService{repo: repo}
}

func (s *PipelineService) CreatePipeline(name string, userID string) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{
		Name:   name,
		UserID: userID,
		Stages: []model.Stage{},
	}
	err := s.repo.CreatePipeline(pipeline)
	return pipeline, err
}

func (s *PipelineService) GetPipeline(id uint, userID string) (*model.Pipeline, error) {
	return s.repo.GetPipelineByIDForUser(id, userID)
}

func (s *PipelineService) GetUserPipelines(userID string) ([]model.Pipeline, error) {
	return s.repo.GetAllPipelines(userID)
}

func (s *PipelineService) UpdatePipeline(id uint, userID string, name string) error {
	return s.repo.UpdatePipeline(id, userID, name)
}

func (s *PipelineService) DeletePipeline(id uint, userID string) error {
	return s.repo.DeletePipeline(id, userID)
}

// Stage services
func (s *PipelineService) CreateStage(name string, pipelineID uint, userID string, position int) (*model.Stage, error) {
	if _, err := s.repo.GetPipelineByIDForUser(pipelineID, userID); err != nil {
		return nil, err
	}

	stage := &model.Stage{
		Name:       name,
		PipelineID: pipelineID,
		Position:   position,
	}
	err := s.repo.CreateStage(stage)
	return stage, err
}

func (s *PipelineService) UpdateStage(stageID uint, pipelineID uint, userID string, name string) error {
	if _, err := s.repo.GetStageByIDForPipeline(stageID, pipelineID, userID); err != nil {
		return err
	}

	stage := &model.Stage{ID: stageID, Name: name}
	return s.repo.UpdateStage(stage)
}

func (s *PipelineService) UpdateStagePosition(stageID uint, pipelineID uint, userID string, position int) error {
	if _, err := s.repo.GetStageByIDForPipeline(stageID, pipelineID, userID); err != nil {
		return err
	}

	return s.repo.UpdateStagePosition(stageID, position)
}

func (s *PipelineService) DeleteStage(stageID uint, pipelineID uint, userID string) error {
	if _, err := s.repo.GetStageByIDForPipeline(stageID, pipelineID, userID); err != nil {
		return err
	}

	return s.repo.DeleteStage(stageID)
}

// Lead services
func (s *PipelineService) CreateLead(
	name string,
	email string,
	phone string,
	pipelineID uint,
	stageID *uint,
	userID string,
) (*model.Lead, error) {
	if stageID == nil {
		return nil, ErrStageIDRequired
	}

	if _, err := s.repo.GetStageByIDForPipeline(*stageID, pipelineID, userID); err != nil {
		return nil, err
	}

	position, err := s.repo.GetNextLeadPosition(*stageID)
	if err != nil {
		return nil, err
	}

	lead := &model.Lead{
		Name:    name,
		Email:   email,
		Phone:   phone,
		StageID: stageID,
		Position: position,
	}
	err = s.repo.CreateLead(lead)
	return lead, err
}

func (s *PipelineService) UpdateLead(leadID uint, pipelineID uint, userID string, name string, email string, phone string) error {
	if _, err := s.repo.GetLeadByIDForPipeline(leadID, pipelineID, userID); err != nil {
		return err
	}

	lead := &model.Lead{
		ID:    leadID,
		Name:  name,
		Email: email,
		Phone: phone,
	}
	return s.repo.UpdateLead(lead)
}

func (s *PipelineService) MoveLeadToStage(leadID uint, pipelineID uint, userID string, stageID *uint, position int) error {
	if _, err := s.repo.GetLeadByIDForPipeline(leadID, pipelineID, userID); err != nil {
		return err
	}

	if stageID != nil {
		if _, err := s.repo.GetStageByIDForPipeline(*stageID, pipelineID, userID); err != nil {
			return err
		}
	}

	return s.repo.MoveLeadToStage(leadID, stageID, position)
}

func (s *PipelineService) DeleteLead(leadID uint, pipelineID uint, userID string) error {
	if _, err := s.repo.GetLeadByIDForPipeline(leadID, pipelineID, userID); err != nil {
		return err
	}

	return s.repo.DeleteLead(leadID)
}
