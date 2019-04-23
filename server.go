package main

import (
	"encoding/json"
	"flag"
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
	/* Unix Domain Socket */
	netProtocol      = "unix"
	socketPath       = "/etc/ssl/certs/smartkey.socket"
	smartkeyURL      = "https://smartkey.io"
	version          = "v1beta1"
	runtime          = "Equinix SmartKey"
	runtimeVersion   = "0.0.1"
	maxRetryTimeout  = 60
	retryIncrement   = 5
)

type CommandArgs struct {
	socketFile string
	configFile string
}

/* KeyManagementServiceServer is a gRPC server. */
type KeyManagementServiceServer struct {
	*grpc.Server
	pathToUnixSocket   string
	providerKeyName    *string
	providerKeyVersion *string
	net.Listener
	config map[string]string
}

/* New method to create instance of KeyManagementServiceServer and initialize the member variables. */
func New(pathToUnixSocketFile string, config map[string]string) (*KeyManagementServiceServer, error) {
	keyManagementServiceServer := new(KeyManagementServiceServer)
	keyManagementServiceServer.pathToUnixSocket = pathToUnixSocketFile
	keyManagementServiceServer.config = config

	log.Println("hello", keyManagementServiceServer.pathToUnixSocket)
	log.Println("yello", keyManagementServiceServer.config)
	return keyManagementServiceServer, nil
}

/* This is a function to parse command line parameters. */
func parseCmd() CommandArgs {
	socketFile := flag.String("socketFile", "", "socket file that gRpc server listens to")
	configFile := flag.String("config", "", "config file location")
	flag.Parse()
	
	if len(*socketFile) == 0 {
		log.Fatal("socketFile parameter not specified")
	}

	if len(*configFile) == 0 {
		log.Fatal("configFile parameter not specified")
        }


        cmdArgs := CommandArgs{
		socketFile:            *socketFile,
		configFile:            *configFile,
	}
	return cmdArgs
}

/* parseConfigFile read file from given path and create dicttionary with properties defined */
func parseConfigFile(configFilePath string) map[string]string {
	file, err := os.Open(configFilePath)
	var config map[string]string
	if err != nil {
		log.Fatal("Unable to open config file", err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Unable to parse config file: %v", err)
	}

	_, isAPIKeyPresent := config["smartkeyApiKey"]
	_, isEnckeyUUIDPresent := config["encryptionKeyUuid"]
	_, isIvPresent := config["iv"]
	_, issocketFilePresent := config["socketFile"]
	_, issmartkeyURLPresent := config["smartkeyURL"]

	/* check mandatory fields are define in config file */
	if isAPIKeyPresent == false {
		log.Fatal("property 'smartkeyApiKey' missing in config file")
	}

	if isEnckeyUUIDPresent == false {
		log.Fatal("property 'encryptionKeyUuid' missing in config file")
	}

	if isIvPresent == false {
		log.Fatal("property 'isIvPresent' missing in config file")
	}

	if issocketFilePresent == false {
		log.Fatal("property 'isIvPresent' missing in config file")
	}

	if issmartkeyURLPresent == false {
		log.Fatal("property 'isIvPresent' missing in config file")
	}
	/* end of check mandatory fields are define in config file */

	return config
}

/* This is the main function. */
func main() {
	/* Parse command line arguments */
	cmdArgs := parseCmd()
	configProperties := parseConfigFile(cmdArgs.configFile)

	var (
                debugListenAddr = flag.String("debug-listen-addr", "127.0.0.1:7901", "HTTP listen address.")
        )

	sigChan := make(chan os.Signal, 1)
	/* Register signal handler for SIGTERM */
	signal.Notify(sigChan, syscall.SIGTERM)

	log.Println("KeyManagementServiceServer service starting...")

	smartkeyServer, err := New(configProperties["socketFile"], configProperties)
	if err != nil {
		log.Fatalf("Failed to start, error: %v", err)
	}

	if err := smartkeyServer.cleanSockFile(); err != nil {
		log.Fatalf("Failed to clean sockfile, error: %v", err)
	}

	listener, err := net.Listen(netProtocol, smartkeyServer.pathToUnixSocket)
	if err != nil {
		log.Fatalf("Failed to start listener, error: %v", err)

		/* Clean the socket file if it exists */
		if err := smartkeyServer.cleanSockFile(); err != nil {
			log.Fatalf("Failed to clean sockfile, error: %v", err)
		}
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

/* Function returns version informatino for the gRPC server. */
func (s *KeyManagementServiceServer) Version(ctx context.Context, request *k8spb.VersionRequest) (*k8spb.VersionResponse, error) {
	log.Println(version)

	return &k8spb.VersionResponse{Version: "v1beta1", RuntimeName: "vault", RuntimeVersion: "0.1.0"}, nil
}

/* This function returns encrypted data. */
func (s *KeyManagementServiceServer) Encrypt(ctx context.Context, request *k8spb.EncryptRequest) (*k8spb.EncryptResponse, error) {

	log.Println("Processing EncryptRequest: ")

	var cihper = []byte(encrypt(s.config, string(request.Plain)))
	return &k8spb.EncryptResponse{Cipher: cihper}, nil
}

/* This function returns decrypted data. */
func (s *KeyManagementServiceServer) Decrypt(ctx context.Context, request *k8spb.DecryptRequest) (*k8spb.DecryptResponse, error) {

	log.Println("Processing DecryptRequest: ")

	var plain = []byte(decrypt(s.config, string(request.Cipher)))
	return &k8spb.DecryptResponse{Plain: plain}, nil
}

/* This function cleans the unix socker created for the gRPC server. */
func (s *KeyManagementServiceServer) cleanSockFile() error {

	log.Println("Cleaning up Socket File: ")
	err := syscall.Unlink(s.config["socketPath"])
	log.Println("Cleaning up Socket File: %v", err)

	return nil
}
