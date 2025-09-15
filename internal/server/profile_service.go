package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/network"
	"github.com/Eitol/NetTestLab/internal/profiles"
)

// ProfileService implements the Connect profile service
type ProfileService struct {
	manager    *profiles.Manager
	controller *network.Controller
}

// NewProfileService creates a new profile service
func NewProfileService(manager *profiles.Manager, controller *network.Controller) *ProfileService {
	return &ProfileService{
		manager:    manager,
		controller: controller,
	}
}

// ListProfiles returns all available profiles
func (s *ProfileService) ListProfiles(ctx context.Context, req *connect.Request[nettestlabv1.ListProfilesRequest]) (*connect.Response[nettestlabv1.ListProfilesResponse], error) {
	profiles := s.manager.ListProfiles(req.Msg.Type, req.Msg.BuiltInOnly)

	return connect.NewResponse(&nettestlabv1.ListProfilesResponse{
		Profiles: profiles,
	}), nil
}

// GetProfile returns a specific profile by name
func (s *ProfileService) GetProfile(ctx context.Context, req *connect.Request[nettestlabv1.GetProfileRequest]) (*connect.Response[nettestlabv1.GetProfileResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("profile name is required"))
	}

	profile, exists := s.manager.GetProfile(req.Msg.Name)
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("profile %s not found", req.Msg.Name))
	}

	return connect.NewResponse(&nettestlabv1.GetProfileResponse{
		Profile: profile,
	}), nil
}

// CreateProfile creates a new custom profile
func (s *ProfileService) CreateProfile(ctx context.Context, req *connect.Request[nettestlabv1.CreateProfileRequest]) (*connect.Response[nettestlabv1.CreateProfileResponse], error) {
	if req.Msg.Profile == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("profile is required"))
	}
	if req.Msg.Profile.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("profile name is required"))
	}

	err := s.manager.CreateProfile(req.Msg.Profile)
	if err != nil {
		return connect.NewResponse(&nettestlabv1.CreateProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	// Get the created profile with metadata
	createdProfile, _ := s.manager.GetProfile(req.Msg.Profile.Name)

	return connect.NewResponse(&nettestlabv1.CreateProfileResponse{
		Success:        true,
		CreatedProfile: createdProfile,
	}), nil
}

// UpdateProfile updates an existing profile
func (s *ProfileService) UpdateProfile(ctx context.Context, req *connect.Request[nettestlabv1.UpdateProfileRequest]) (*connect.Response[nettestlabv1.UpdateProfileResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("profile name is required"))
	}
	if req.Msg.Profile == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("profile is required"))
	}

	err := s.manager.UpdateProfile(req.Msg.Name, req.Msg.Profile)
	if err != nil {
		return connect.NewResponse(&nettestlabv1.UpdateProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	// Get the updated profile
	updatedProfile, _ := s.manager.GetProfile(req.Msg.Name)

	return connect.NewResponse(&nettestlabv1.UpdateProfileResponse{
		Success:        true,
		UpdatedProfile: updatedProfile,
	}), nil
}

// DeleteProfile deletes a profile
func (s *ProfileService) DeleteProfile(ctx context.Context, req *connect.Request[nettestlabv1.DeleteProfileRequest]) (*connect.Response[nettestlabv1.DeleteProfileResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("profile name is required"))
	}

	err := s.manager.DeleteProfile(req.Msg.Name)
	if err != nil {
		return connect.NewResponse(&nettestlabv1.DeleteProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&nettestlabv1.DeleteProfileResponse{
		Success: true,
	}), nil
}

// ApplyProfile applies a profile to an interface
func (s *ProfileService) ApplyProfile(ctx context.Context, req *connect.Request[nettestlabv1.ApplyProfileRequest]) (*connect.Response[nettestlabv1.ApplyProfileResponse], error) {
	if req.Msg.ProfileName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("profile name is required"))
	}
	if req.Msg.Interface == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("interface name is required"))
	}

	// Get the profile
	profile, exists := s.manager.GetProfile(req.Msg.ProfileName)
	if !exists {
		return connect.NewResponse(&nettestlabv1.ApplyProfileResponse{
			Success:      false,
			ErrorMessage: "profile not found",
		}), nil
	}

	// Set default direction if not specified
	direction := req.Msg.Direction
	if direction == nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_UNSPECIFIED {
		direction = nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_BOTH
	}

	// Apply the profile's conditions
	err := s.controller.ApplyConditions(req.Msg.Interface, profile.Conditions, direction)
	if err != nil {
		return connect.NewResponse(&nettestlabv1.ApplyProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	// Store the applied profile name in the controller
	s.controller.SetAppliedProfile(req.Msg.Interface, req.Msg.ProfileName)

	return connect.NewResponse(&nettestlabv1.ApplyProfileResponse{
		Success:           true,
		AppliedConditions: profile.Conditions,
	}), nil
}
