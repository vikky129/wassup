package bl

import (
	"context"
	"fmt"
	"math/rand"
	"personal/wassup/middleware"
	"personal/wassup/spec"
	"time"
)

// Soferi
// Ultra D3

func (wh *WassupHandler) SendOTP(ctx context.Context, otpRequest *spec.SendOtpRequest) (*spec.SendOtpResponse, error) {
	otp := rand.Int63n(10000)
	err := wh.redisHandler.Set("otp:"+otpRequest.PhoneNumber, otp)
	if err != nil {
		return nil, err
	}

	return &spec.SendOtpResponse{
		OTP: otp,
	}, nil
}

func (wh *WassupHandler) VerifyOTP(ctx context.Context, otpRequest *spec.VerifyOtpRequest) (*spec.VerifyOtpResponse, error) {
	otp, err := wh.redisHandler.Get("otp:" + otpRequest.PhoneNumber)
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

	userID, err := wh.dbHandler.UserHandler.CreateUser(ctx, userObj)
	if err != nil {
		return nil, err
	}

	jwt, err := middleware.GenerateJWT(userID)
	if err != nil {
		return nil, err
	}

	return &spec.VerifyOtpResponse{
		Success: true,
		UserID:  userID,
		Token:   jwt,
	}, nil
}

func (wh *WassupHandler) SetProfile(ctx context.Context, setProfileRequest *spec.SetProfileRequest) (*spec.SetProfileResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	userObj, err := wh.dbHandler.UserHandler.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userObj.Name = setProfileRequest.Name
	userObj.Bio = setProfileRequest.Bio
	userObj.UpdatedAt = time.Now().Unix()

	err = wh.dbHandler.UserHandler.UpdateUser(ctx, userObj)
	if err != nil {
		return nil, err
	}

	return &spec.SetProfileResponse{
		Result: "Profile updated successfully",
	}, nil
}

func (wh *WassupHandler) UploadProfileImage(ctx context.Context, userID, filePath string) error {

	userObj, err := wh.dbHandler.UserHandler.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	userObj.PhotoURL = filePath
	userObj.UpdatedAt = time.Now().Unix()

	err = wh.dbHandler.UserHandler.UpdateUser(ctx, userObj)
	if err != nil {
		return err
	}

	return nil
}
