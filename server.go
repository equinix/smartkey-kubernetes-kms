// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"

	k8spb "smartkey/v1beta1"
)

const (
	// Unix Domain Socket
	netProtocol      = "unix"
	socketPath       = "/opt/smartkey.socket"
	version          = "v1beta1"
	runtime          = "Equinix SmartKey"
	runtimeVersion   = "0.0.1"
	maxRetryTimeout  = 60
	retryIncrement   = 5
)

type CommandArgs struct {
	socketFile            string
	smartkeyConfig           string
}

// KeyManagementServiceServer is a gRPC server.
type KeyManagementServiceServer struct {
	*grpc.Server
	pathToUnixSocket   string
	providerKeyName    *string
	providerKeyVersion *string
	net.Listener
	configFilePath string
}

func New(pathToUnixSocketFile string, configFilePath string) (*KeyManagementServiceServer, error) {
	keyManagementServiceServer := new(KeyManagementServiceServer)
	keyManagementServiceServer.pathToUnixSocket = pathToUnixSocketFile
	keyManagementServiceServer.configFilePath = configFilePath

	log.Println(keyManagementServiceServer.pathToUnixSocket)
	log.Println(keyManagementServiceServer.configFilePath)
	return keyManagementServiceServer, nil
}

func parseCmd() CommandArgs {
	socketFile := flag.String("socketFile", "", "socket file that gRpc server listens to")
	smartkeyConfig := flag.String("smartkeyConfig", "", "smartkey config file location")
	flag.Parse()
	
	if len(*socketFile) == 0 {
		log.Fatal("socketFile parameter not specified")
	}

	if len(*smartkeyConfig) == 0 {
		log.Fatal("smartkeyConfig parameter not specified")
        }


        cmdArgs := CommandArgs{
		socketFile:            *socketFile,
		smartkeyConfig:           *smartkeyConfig,
	}
	return cmdArgs
}

func main() {
	/* Parse command line arguments */
	cmdArgs := parseCmd()

	var (
                debugListenAddr = flag.String("debug-listen-addr", "127.0.0.1:7901", "HTTP listen address.")
        )

	sigChan := make(chan os.Signal, 1)
	/* Register signal handler for SIGTERM */
	signal.Notify(sigChan, syscall.SIGTERM)

	log.Println("KeyManagementServiceServer service starting...")
	smartkeyServer, err := New(socketPath, cmdArgs.smartkeyConfig)
	if err != nil {
		log.Fatalf("Failed to start, error: %v", err)
	}

	if err := smartkeyServer.cleanSockFile(); err != nil {
		log.Fatalf("Failed to clean sockfile, error: %v", err)
	}

	listener, err := net.Listen(netProtocol, smartkeyServer.pathToUnixSocket)
	if err != nil {
		log.Fatalf("Failed to start listener, error: %v", err)
	}
	smartkeyServer.Listener = listener

	server := grpc.NewServer()
	k8spb.RegisterKeyManagementServiceServer(server, smartkeyServer)
	smartkeyServer.Server = server

	go server.Serve(listener)
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) { return true, true }
	log.Println("KeyManagementServiceServer service started successfully.")

	go func() {
		for {
			smartkeyServer := <-sigChan
			if smartkeyServer == syscall.SIGTERM {
				log.Println("force stop")
				log.Println("Shutting down gRPC service...")
				server.GracefulStop()
				os.Exit(0)
			}
		}
	}()
	log.Fatal(http.ListenAndServe(*debugListenAddr, nil))
}

func (s *KeyManagementServiceServer) Version(ctx context.Context, request *k8spb.VersionRequest) (*k8spb.VersionResponse, error) {
	fmt.Println(version)

	return &k8spb.VersionResponse{Version: "v1beta1", RuntimeName: "vault", RuntimeVersion: "0.1.0"}, nil
}

func (s *KeyManagementServiceServer) Encrypt(ctx context.Context, request *k8spb.EncryptRequest) (*k8spb.EncryptResponse, error) {

	log.Println("Processing EncryptRequest: ")

	return &k8spb.EncryptResponse{Cipher: request.Plain}, nil
}

func (s *KeyManagementServiceServer) Decrypt(ctx context.Context, request *k8spb.DecryptRequest) (*k8spb.DecryptResponse, error) {

	log.Println("Processing DecryptRequest: ")

	return &k8spb.DecryptResponse{Plain: request.Cipher}, nil
}

func (s *KeyManagementServiceServer) cleanSockFile() error {

	log.Println("Cleaning up Socket File: ")

	return nil
}
