package gRPCProto

import (
	"FunGoLoginCenter/database"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type UserService struct {
	UnimplementedUserServiceServer
}

func (u *UserService) GetUserInfo(c context.Context, req *UserInfoReq) (*UserInfoRsp, error) {
	userInfo := database.UserInfo{}
	if err := database.Db.Where("uid = ?", req.Uid).First(&userInfo).Error; err != nil {
		return nil, err
	} else {
		res := UserInfoRsp{}
		res.Uid = userInfo.UID
		res.UserEmail = userInfo.UserEmail
		res.Name = userInfo.UserName
		return &res, nil
	}
}

type CheckLoginService struct {
	UnimplementedCheckLoginServiceServer
}

func (c *CheckLoginService) CheckLogin(context context.Context, req *CheckLoginReq) (*CheckLoginRsp, error) {
	m, err := database.Client.Do(context, database.Client.B().Hmget().Key(req.Key).Field("uid", "name").Build()).ToArray()
	if err != nil {
		return nil, nil
	}
	res := CheckLoginRsp{}
	mUID := m[0]
	id, _ := mUID.AsInt64()
	mName := m[1]
	res.Uid = uint32(id)
	res.UserName, _ = mName.ToString()
	database.Client.Do(context, database.Client.B().Del().Key(req.Key).Build())
	return &res, nil
}

func InitGRPCServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7777))
	if err != nil {
		panic(err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	//注册服务
	RegisterUserServiceServer(grpcServer, &UserService{})
	RegisterCheckLoginServiceServer(grpcServer, &CheckLoginService{})
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
