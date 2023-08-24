package handler

import (
	"context"
	"fmt"
	"api_gateway/kitex_gen/user"
	"api_gateway/kitex_gen/user/userservice"
	"net/http"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/client"
)

type UserImpl struct {
	client userservice.Client
}

type UserRegisterResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	UserID     int64  `json:"user_id"`
	Token      string `json:"token"`
}

func NewUserImpl() *UserImpl {
	c, err := userservice.NewClient("user", client.WithHostPorts("127.0.0.1:9990"))
	if err != nil {
		panic(fmt.Sprintf("create user client error: %v", err))
	}
	return &UserImpl{client: c} // 指定下游的ip，高级用法可以使用resolver去调用服务注册中心
}



// Login: Post请求
func (u *UserImpl) Register(ctx context.Context, c *app.RequestContext) {
	username, password := c.PostForm("username"), c.PostForm("password")
	lr, err := u.client.Register(ctx, &user.DouyinUserRegisterRequest{Username: username, Password: password})
	if err != nil {
		response := UserRegisterResponse{
			StatusMsg:  "fail",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	if lr.GetStatusCode() == -1 {
		response := UserRegisterResponse{
			StatusMsg:  lr.GetStatusMsg(),
		}
		
		c.JSON(http.StatusInternalServerError,response )
		return
	}
	response := UserRegisterResponse{
		StatusCode: lr.GetStatusCode(), // 成功状态码
		StatusMsg:  lr.GetStatusMsg(),
		UserID:     lr.GetUserId(),
		Token:      lr.GetToken(),
	}

	c.JSON(http.StatusOK, response)
}

// // LogOut: Post请求, 同上
// func (u *UserImpl) LogOut(ctx context.Context, c *app.RequestContext) {
// 	username := c.PostForm("username")
// 	lor, err := u.client.LogOut(ctx, &biz.LogoutRequest{UserToken: username})
// 	if err != nil {
// 		c.JSON(200, data{
// 			"msg":  err.Error(),
// 			"data": "",
// 			"code": -1,
// 		})
// 		return
// 	}
// 	if lor.GetBase().GetCode() == -1 {
// 		c.JSON(200, data{
// 			"msg":  lor.GetBase().GetMsg(),
// 			"data": "",
// 			"code": lor.GetBase().GetCode(),
// 		})
// 		return
// 	}
// 	c.JSON(200, data{
// 		"msg":  lor.GetBase().GetMsg(),
// 		"data": "",
// 		"code": lor.GetBase().GetCode(),
// 	})
// }
