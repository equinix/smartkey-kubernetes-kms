package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
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

	k8spb "smartkey-kubernetes-kms/v1beta1"
)

const (
	/* Unix Domain Socket */
	netProtocol     = "unix"
	version         = "v1beta1"
	runtime         = "Equinix SmartKey"
	runtimeVersion  = "0.1.0"
	maxRetryTimeout = 60
	retryIncrement  = 5
)

/*CommandArgs ...*/
type CommandArgs struct {
	socketFile string
	configFile string
}

/*KeyManagementServiceServer is a gRPC server. */
type KeyManagementServiceServer struct {
	*grpc.Server
	pathToUnixSocket   string
	providerKeyName    *string
	providerKeyVersion *string
	net.Listener
	config map[string]string
}

/*New creates instance of KeyManagementServiceServer and initialize the member variables. */
func New(pathToUnixSocketFile string, config map[string]string) (*KeyManagementServiceServer, error) {
	keyManagementServiceServer := new(KeyManagementServiceServer)
	keyManagementServiceServer.pathToUnixSocket = pathToUnixSocketFile
	keyManagementServiceServer.config = config

	return keyManagementServiceServer, nil
}

/* This is a function to parse command line parameters. */
func parseCmd() (CommandArgs, error) {
	socketFile := flag.String("socketFile", "", "socket file that gRpc server listens to")
	configFile := flag.String("config", "", "config file location")
	flag.Parse()
	var cmdArgs CommandArgs

	if len(*socketFile) == 0 {
		return cmdArgs, errors.New("socketFile parameter not specified")
	}

	if len(*configFile) == 0 {
		return cmdArgs, errors.New("configFile parameter not specified")
	}

	cmdArgs = CommandArgs{
		socketFile: *socketFile,
		configFile: *configFile,
	}
	return cmdArgs, nil
}

/* parseConfigFile read file from given path and create dictionary with properties defined */
func parseConfigFile(configFilePath string) (map[string]string, error) {
	file, err := os.Open(configFilePath)
	var config map[string]string
	if err != nil {
		return nil, errors.New("Unable to open config file " + configFilePath)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, errors.New("Unable to parse config file " + configFilePath)
	}

	_, isAPIKeyPresent := config["smartkeyApiKey"]
	_, isEnckeyUUIDPresent := config["encryptionKeyUuid"]
	_, isIvPresent := config["iv"]
	_, issocketFilePresent := config["socketFile"]
	_, issmartkeyURLPresent := config["smartkeyURL"]

	/* check mandatory fields are define in config file */
	if isAPIKeyPresent == false {
		return nil, errors.New("property 'smartkeyApiKey' missing in config file " + configFilePath)
	}

	if isEnckeyUUIDPresent == false {
		return nil, errors.New("property 'encryptionKeyUuid' missing in config file " + configFilePath)
	}

	if isIvPresent == false {
		return nil, errors.New("property 'isIvPresent' missing in config file " + configFilePath)
	}

	if issocketFilePresent == false {
		return nil, errors.New("property 'socketFile' missing in config file " + configFilePath)
	}

	if issmartkeyURLPresent == false {
		return nil, errors.New("property 'smartkeyURL' missing in config file " + configFilePath)
	}
	/* end of check mandatory fields are define in config file */

	/* validate Api key and AES key */
	_, err = auth(config)
	if err != nil {
		return nil, errors.New("property 'smartkeyApiKey' is invalid in config file " + configFilePath)
	}

	_, err = validateKey(config)
	if err != nil {
		return nil, errors.New("property 'encryptionKeyUuid' is invalid in config file " + configFilePath)
	}

	decodeIv, decodeIvErr := base64.StdEncoding.DecodeString(config["iv"])
	if decodeIvErr != nil {
		return nil, errors.New("property 'iv' has an invalid format in config file " + configFilePath)
	}
	if (len(decodeIv)) != 16 {
		return nil, errors.New("property 'iv' has an invalid format in config file " + configFilePath)
	}

	/* end of Api key and AES key */

	return config, nil
}

/* This is the main function. */
func main() {
	/* Parse command line arguments */
	cmdArgs, commandErr := parseCmd()
	if commandErr != nil {
		log.Fatal(commandErr)
	}

	configProperties, fileErr := parseConfigFile(cmdArgs.configFile)
	if fileErr != nil {
		log.Fatal(fileErr)
	}

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

/*Version returns version informatino for the gRPC server. */
func (s *KeyManagementServiceServer) Version(ctx context.Context, request *k8spb.VersionRequest) (*k8spb.VersionResponse, error) {
	log.Println(version)

	return &k8spb.VersionResponse{Version: version, RuntimeName: "vault", RuntimeVersion: runtimeVersion}, nil
}

/*Encrypt function returns encrypted data. */
func (s *KeyManagementServiceServer) Encrypt(ctx context.Context, request *k8spb.EncryptRequest) (*k8spb.EncryptResponse, error) {

	log.Println("Processing EncryptRequest: ")

	response, err := encrypt(s.config, string(request.Plain))
	return &k8spb.EncryptResponse{Cipher: []byte(response)}, err
}

/*Decrypt function returns decrypted data. */
func (s *KeyManagementServiceServer) Decrypt(ctx context.Context, request *k8spb.DecryptRequest) (*k8spb.DecryptResponse, error) {

	log.Println("Processing DecryptRequest: ")

	response, err := decrypt(s.config, string(request.Cipher))
	return &k8spb.DecryptResponse{Plain: []byte(response)}, err
}

/*cleanSockFile function cleans the unix socker created for the gRPC server. */
func (s *KeyManagementServiceServer) cleanSockFile() error {

	log.Println("Cleaning up Socket File: ")
	err := syscall.Unlink(s.config["socketPath"])
	log.Println("Cleaning up Socket File:", err)

	return nil
}
