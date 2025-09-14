package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	nettestlabv1 "github.com/Eitol/NetTestLab/api/nettestlab/v1"
	"github.com/Eitol/NetTestLab/internal/network"
	"github.com/Eitol/NetTestLab/internal/profiles"
)

// ProfileService implements the gRPC profile service
type ProfileService struct {
	nettestlabv1.UnimplementedProfileServiceServer
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
func (s *ProfileService) ListProfiles(ctx context.Context, req *nettestlabv1.ListProfilesRequest) (*nettestlabv1.ListProfilesResponse, error) {
	profiles := s.manager.ListProfiles(req.Type, req.BuiltInOnly)

	return &nettestlabv1.ListProfilesResponse{
		Profiles: profiles,
	}, nil
}

// GetProfile returns a specific profile by name
func (s *ProfileService) GetProfile(ctx context.Context, req *nettestlabv1.GetProfileRequest) (*nettestlabv1.GetProfileResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "profile name is required")
	}

	profile, exists := s.manager.GetProfile(req.Name)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "profile %s not found", req.Name)
	}

	return &nettestlabv1.GetProfileResponse{
		Profile: profile,
	}, nil
}

// CreateProfile creates a new custom profile
func (s *ProfileService) CreateProfile(ctx context.Context, req *nettestlabv1.CreateProfileRequest) (*nettestlabv1.CreateProfileResponse, error) {
	if req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}
	if req.Profile.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "profile name is required")
	}

	err := s.manager.CreateProfile(req.Profile)
	if err != nil {
		return &nettestlabv1.CreateProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Get the created profile with metadata
	createdProfile, _ := s.manager.GetProfile(req.Profile.Name)

	return &nettestlabv1.CreateProfileResponse{
		Success:        true,
		CreatedProfile: createdProfile,
	}, nil
}

// UpdateProfile updates an existing profile
func (s *ProfileService) UpdateProfile(ctx context.Context, req *nettestlabv1.UpdateProfileRequest) (*nettestlabv1.UpdateProfileResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "profile name is required")
	}
	if req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}

	err := s.manager.UpdateProfile(req.Name, req.Profile)
	if err != nil {
		return &nettestlabv1.UpdateProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Get the updated profile
	updatedProfile, _ := s.manager.GetProfile(req.Name)

	return &nettestlabv1.UpdateProfileResponse{
		Success:        true,
		UpdatedProfile: updatedProfile,
	}, nil
}

// DeleteProfile deletes a profile
func (s *ProfileService) DeleteProfile(ctx context.Context, req *nettestlabv1.DeleteProfileRequest) (*nettestlabv1.DeleteProfileResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "profile name is required")
	}

	err := s.manager.DeleteProfile(req.Name)
	if err != nil {
		return &nettestlabv1.DeleteProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &nettestlabv1.DeleteProfileResponse{
		Success: true,
	}, nil
}

// ApplyProfile applies a profile to an interface
func (s *ProfileService) ApplyProfile(ctx context.Context, req *nettestlabv1.ApplyProfileRequest) (*nettestlabv1.ApplyProfileResponse, error) {
	if req.ProfileName == "" {
		return nil, status.Error(codes.InvalidArgument, "profile name is required")
	}
	if req.Interface == "" {
		return nil, status.Error(codes.InvalidArgument, "interface name is required")
	}

	// Get the profile
	profile, exists := s.manager.GetProfile(req.ProfileName)
	if !exists {
		return &nettestlabv1.ApplyProfileResponse{
			Success:      false,
			ErrorMessage: "profile not found",
		}, nil
	}

	// Set default direction if not specified
	direction := req.Direction
	if direction == nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_UNSPECIFIED {
		direction = nettestlabv1.TrafficDirection_TRAFFIC_DIRECTION_BOTH
	}

	// Apply the profile's conditions
	err := s.controller.ApplyConditions(req.Interface, profile.Conditions, direction)
	if err != nil {
		return &nettestlabv1.ApplyProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &nettestlabv1.ApplyProfileResponse{
		Success:           true,
		AppliedConditions: profile.Conditions,
	}, nil
}
