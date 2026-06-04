package bl

import (
	"context"
	"fmt"
	"math/rand"
	"personal/wassup/spec"
	"time"
)

// Soferi
// Ultra D3

func (svc *authService) SendOTP(ctx context.Context, otpRequest *spec.SendOtpRequest) (*spec.SendOtpResponse, error) {
	otp := rand.Int63n(10000)
	err := svc.redisHandler.Set("otp:"+otpRequest.PhoneNumber, otp)
	if err != nil {
		return nil, err
	}

	return &spec.SendOtpResponse{
		OTP: otp,
	}, nil
}

func (svc *authService) VerifyOTP(ctx context.Context, otpRequest *spec.VerifyOtpRequest) (*spec.VerifyOtpResponse, error) {
	otp, err := svc.redisHandler.Get("otp:" + otpRequest.PhoneNumber)
	if err != nil {
		return nil, err
	}

	if otp != otpRequest.Otp {
		return &spec.VerifyOtpResponse{
			Success: false,
		}, nil
	}

	userObj := &spec.User{
		PhoneNumber: otpRequest.PhoneNumber,
		IsVerified:  true,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	userID, err := svc.dbHandler.CreateUser(ctx, userObj)
	if err != nil {
		return nil, err
	}

	jwt, err := svc.tokenService.GenerateToken(userID)
	if err != nil {
		return nil, err
	}

	return &spec.VerifyOtpResponse{
		Success: true,
		UserID:  userID,
		Token:   jwt,
	}, nil
}

func (svc *userService) SetProfile(ctx context.Context, setProfileRequest *spec.SetProfileRequest) (*spec.SetProfileResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	userObj, err := svc.dbHandler.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userObj.Name = setProfileRequest.Name
	userObj.Bio = setProfileRequest.Bio
	userObj.UpdatedAt = time.Now().Unix()

	err = svc.dbHandler.UpdateUser(ctx, userObj)
	if err != nil {
		return nil, err
	}

	return &spec.SetProfileResponse{
		Result: "Profile updated successfully",
	}, nil
}

func (svc *userService) UploadProfileImage(ctx context.Context, userID, filePath string) error {

	userObj, err := svc.dbHandler.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	userObj.PhotoURL = filePath
	userObj.UpdatedAt = time.Now().Unix()

	err = svc.dbHandler.UpdateUser(ctx, userObj)
	if err != nil {
		return err
	}

	return nil
}
