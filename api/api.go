package api

import (
	"bogolang/config"
	"bogolang/server/pb"
	"context"
	"fmt"
)

type Server struct {
	pb.BogolangServer
}

func (s *Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	fmt.Println("CreateUser:ENTER")
	err := config.DB.Create(in).Error
	if err != nil {
		return &pb.CreateUserResponse{}, err
	}
	return &pb.CreateUserResponse{
		Status:  true,
		Message: "success",
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	fmt.Println("DeleteUser:ENTER")
	err := config.DB.Delete("id = ?", in.GetUserId()).Error
	if err != nil {
		return &pb.DeleteUserResponse{
			Status: false,
		}, err
	}
	return &pb.DeleteUserResponse{
		Status: true,
	}, nil
}
func (s *Server) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	fmt.Println("UpdateUser:ENTER")
	err := config.DB.Update("name = ? ", in.GetName()).Error
	if err != nil {
		return &pb.UpdateUserResponse{
			Status: false,
		}, err
	}
	return &pb.UpdateUserResponse{
		Status:  true,
		Message: "success",
	}, nil
}
