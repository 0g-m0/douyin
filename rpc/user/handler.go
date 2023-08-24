package main

import (
	"context"
	user "user/kitex_gen/user"
	"user/database"
	"user/database/models"
	"golang.org/x/crypto/bcrypt"
	
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct{}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, request *user.DouyinUserRegisterRequest) (resp *user.DouyinUserRegisterResponse, err error) {
	// TODO: Your code here...
	resp = &user.DouyinUserRegisterResponse{}

	if request.GetUsername() == "" || request.GetPassword() == "" {
		resp = &user.DouyinUserRegisterResponse{StatusCode: -1, StatusMsg: "username or password is empty"}
		return
	}

	hashPassword, err := hashPassword(request.GetPassword())
	if err != nil {
		resp = &user.DouyinUserRegisterResponse{StatusCode: -1, StatusMsg: "密码加密失败"}
		return
	}

	// 创建用户数据模型
	newuser := models.User{
		Name:     request.GetUsername(),
		Password: hashPassword,
	}

	// 验证用户名是否已经存在
	err = database.DB.Table("user").Where("name = ?", newuser.Name).First(&newuser).Error
	if err == nil {
		resp = &user.DouyinUserRegisterResponse{StatusCode: -1, StatusMsg: "注册用户名已经存在"}
		return
	}

	//保存用户数据到数据库
	err = database.DB.Table("user").Create(&newuser).Error
	if err != nil {
		resp = &user.DouyinUserRegisterResponse{StatusCode: -1, StatusMsg: "注册信息写入数据库时创建用户失败"}
		return
	}



	resp = &user.DouyinUserRegisterResponse{StatusCode: 0, StatusMsg: "注册用户成功",UserId: newuser.ID ,Token: "DeFault_Token_Wait_For_Inple"}
	return
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, request *user.DouyinUserLoginRequest) (resp *user.DouyinUserLoginResponse, err error) {
	// TODO: Your code here...
	return
}

// GetUserById implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserById(ctx context.Context, request *user.DouyinUserRequest) (resp *user.DouyinUserResponse, err error) {
	// TODO: Your code here...
	return
}


// 加密用户密码
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}