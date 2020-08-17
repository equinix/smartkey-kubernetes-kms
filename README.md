# SmartKey Kubernetes KMS Plugin

This project allows you to use SmartKey as a Key Management Service (KMS) provider for Kubernetes. Refer to [https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/](https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/) to understand how Kubernetes works with an external KMS.

Before you start:
1. GoLang must be installed in the server you want to build the project on/create installer (Running the installer does not require GoLang).
2. Existing Kubernetes cluster must be configured to run "SmartKey Kubernetes KMS Plugin".
3. In order to build/run code or to create the installer, git pull this project into your go directory (eg: ~/go/src/)



## Getting Started (Developer)
Checkout a current version of the project.

##### To test, build or clean the project (from within the project)
This is only for developers who want to build a binary of the gRPC server manually or to execute unit test cases.
- To test code, run the below command

		$ make test
- To build code, run the below command 

		$ make build
- To clean workspace, run the below command

		$ make clean

##### To generate smartkey-kms binary using build command
For developers who want to run the gRPC server directly without using an installer, follow these steps.
  - Execute the following commands to run the binary.

		make build (to build your code and get smartkey-kms)
		sudo ./smartkey-kms --socketFile <sock-file-path> --config <config-file->
		
       - \<sock-file-path>:  Path where you want to create your unix socket file eg: /etc/smartkey/smartkey.socket
       - \<config-file>: Path to your config file. (eg. conf/smartkey-grpc.conf)
       - **Note**: \<sock-file-path> must already exist (eg. /etc/smartkey). If not, please create before running server.

##### To create a Debian installer from plugin binary
  - Install these tools

		sudo apt-get install dh-make
		sudo apt-get install devscripts build-essential lintian
  - Execute the script "create_installer.sh". It will create .deb file in one level above your PWD

## Installation
Instead of building code manually  (as mentioned in **Getting started (Developer)** section), we can install the plugin using the compiled .deb file created in the previous step **To create a Debian installer from plugin binary**.

#### To install and configure the plugin from Debian installer
  - Execute this command to install the package

		sudo dpkg -i smartkey-kmsplugin_1.0-1_amd64.deb
  - The above command will only install the plugin binary in your machine but will not start it.
  - **Note**: We will need to do configure our config files before starting service.
  - Update the configuration at "/etc/smartkey/smartkey-grpc.conf". 
  
    Equinix Smartkey URLs:
    
    - North America: https://amer.smartkey.io/
    - European Union: https://eu.smartkey.io/
    - United Kingdom: https://uk.smartkey.io/
    - Asia Pacific: https://apac.smartkey.io/
    - Australia: https://au.smartkey.io/
    

	**Sample "smartkey-grpc.conf" file**

		{
	      "smartkeyApiKey": "<smartkey-api-key>",
		  "encryptionKeyUuid": "<uuid-for-aes-encryption-key-in-SmartKey>",
		  "iv": "<Initialization vector for your AES algo as per your key size (128, 192, 256)>",
		  "socketFile": "<path-to-your-sock-file>",
		  "smartkeyURL": "<smartkey-url>"
		}
  - Execute the following command to run the plugin gRPC server 
    
	    sudo service smartkey-grpc start &
  - Verify if the service was started correctly using this command
    
	    sudo service smartkey-grpc status

## Configuring the api-server to use "SmartKey Kubernetes KMS Plugin"
For both methods, 
1) running plugin manually or 
2) running the "smartkey-grpc" service created by the installer, 

we need to follow these steps.

1. Open "/etc/kubernetes/manifests/kube-apiserver.yaml" file and add the below configuration.

	    - spec.containers.command.kube-apiserver
	        --encryption-provider-config=/etc/smartkey/smartkey.yaml

2. The installer will automatically create "smartkey.yaml" and copy it to the desired (/etc/smartkey/) location with pre-populated configuration.

3. Add "volumemount" and "volume host path" to "kube-apiserver.yaml". 
Without following configurations, api-server will not be able to read "smartkey.yaml" and "smartkey.sock" files.


	    =====
	    volumeMounts:
	       - mountPath: /etc/smartkey
	           name: smartkey-kms
	           readOnly: true
	    ...
	    ...
	    volumes:
	       - hostPath:
	           path: /etc/smartkey
	           type: DirectoryOrCreate
	         name: smartkey-kms
	    ====
    
	**Installer will not add these configurations. They need to be manually added to "kube-apiserver.yaml".

4. Save "kube-apiserver.yaml" and exit. api-server will now detect changes in "kube-apiserver.yaml" file and restart.
    
## Verifying if plugin is configured correctly
 - Create a new secret and you should be able to see logs in the plugin service logs. 
 - Use this command to see logs
		
		journalctl -xe
 - To create a new secret which will be encrypted by our plugin, use this command
        
        kubectl create secret generic install3 -n default --from-literal=es-engkey=equinixdata
 - To encrypt all of your existing secrets, use this command
	        
        kubectl get secrets --all-namespaces -o json | kubectl replace -f -
    
## Troubleshooting
1. Permission denied error while trying to encrypt old secrets using below command
        Create new secret and you should be able to see logs in plugin service logs. 
        
		$ kubectl get secrets --all-namespaces -o json | kubectl replace -f -
        
	Fix: Make sure that kubernetes config directory has the same permissions as kubernetes config file.

		$ mkdir -p $HOME/.kube
		$ sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
		$ sudo chown $(id -u):$(id -g) $HOME/.kube/config 
        
	Add change permissions on $HOME/.kube/ directory.

		$ sudo chown $(id -u):$(id -g) $HOME/.kube/
2. For error "go get: no install location for directory" execute below commands.

        export GOBIN=$HOME/work/bin
        export GOPATH=$HOME/go
3. In order to uninstall debian package use below command.

        sudo dpkg -P smartkey-kmsplugin

## Support email
For any queries, contact ES-ENG-SECURITY <ES-ENG-SECURITY@equinix.com>