package main

import (
	"bogolang/config"
	"bogolang/server/pb"
	"context"
	"encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

const serviceName = "bogolang"
const defaultPort = 9090

var s *grpc.Server

func main() {
	app := cli.NewApp()
	app.Name = "bogolang"
	app.Usage = "bogolang service"
	app.Commands = []cli.Command{
		grpcServerCommand(),
		gateWayServerCmd(),
		grpcGatewayServerCmd(),
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func grpcServerCommand() cli.Command {
	return cli.Command{
		Name:  "grpc-server",
		Usage: "bogolang service",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "port",
				Value: defaultPort,
			},
		},
		Action: func(c *cli.Context) error {
			port := c.Int("port")
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)

			config.InitConnection()

			go func() {
				if err := grpcServer(port, ch); err != nil {
					fmt.Println(err)
				}
			}()

			<-ch
			s.Stop()
			return nil
		},
	}
}

func grpcServer(port int, _ chan os.Signal) error {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s = grpc.NewServer()
	type server struct {
		pb.UnimplementedBogolangServer
	}
	pb.RegisterBogolangServer(s, server{})
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

	return s.Serve(list)
}

func gateWayServerCmd() cli.Command {
	return cli.Command{
		Name:  "gw-server",
		Usage: "bogolang service",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "port",
				Value: 3000,
			},
			cli.StringFlag{
				Name:  "grpc-endpoint",
				Value: ":" + fmt.Sprint(defaultPort),
			},
		},
		Action: func(c *cli.Context) error {
			port, grpcEndpoint := c.Int("port"), c.String("grpc-endpoint")

			go func() {
				if err := httpGateWayServer(port, grpcEndpoint); err != nil {
					fmt.Println(err)

				}
			}()

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			<-ch
			return nil

		},
	}
}

func grpcGatewayServerCmd() cli.Command {
	return cli.Command{
		Name:  "grpc-gw-server",
		Usage: "bogolang service",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "port1",
				Value: defaultPort,
			},
			cli.IntFlag{
				Name:  "port2",
				Value: 3000,
			},
			cli.StringFlag{
				Name:  "grpc-endpoint",
				Value: ":" + fmt.Sprint(defaultPort),
			},
		},
		Action: func(c *cli.Context) error {
			rpcPort, httpPort, grpcEndpoint := c.Int("port1"), c.Int("port2"), c.String("grpc-endpoint")

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			config.InitConnection()

			go func() {
				if err := grpcServer(rpcPort, ch); err != nil {
					fmt.Println(err)
				}
			}()

			go func() {
				if err := httpGateWayServer(httpPort, grpcEndpoint); err != nil {
				}
			}()

			<-ch

			s.Stop()
			return nil
		},
	}
}

func httpGateWayServer(port int, grpcEndPoint string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := grpc.Dial(
		grpcEndPoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	rmux := runtime.NewServeMux(runtime.WithErrorHandler(CustomHTTPError))

	client := pb.NewBogolangClient(conn)
	err = pb.RegisterBogolangHandlerClient(ctx, rmux, client)
	if err != nil {
		return err
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), rmux)
}

func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	custErr := status.Convert(err)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(runtime.HTTPStatusFromCode(custErr.Code()))

	body := &ErrorBodyResponse{
		Error:   true,
		Code:    uint32(runtime.HTTPStatusFromCode(custErr.Code())),
		Message: custErr.Message(),
	}

	jErr := json.NewEncoder(w).Encode(body)

	if jErr != nil {
		w.Write([]byte(fallback))
	}
}

type ErrorBodyResponse struct {
	Error   bool   `json:"error"`
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}
