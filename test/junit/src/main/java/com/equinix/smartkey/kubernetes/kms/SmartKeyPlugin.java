package com.equinix.smartkey.kubernetes.kms;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;

import java.io.FileInputStream;
import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Properties;
import java.util.Map.Entry;

import org.apache.log4j.Logger;
import org.junit.After;
import org.junit.Before;
import org.junit.FixMethodOrder;
import org.junit.Test;
import org.junit.runners.MethodSorters;
import org.yaml.snakeyaml.DumperOptions;
import org.yaml.snakeyaml.Yaml;

import com.jcraft.jsch.JSchException;

@FixMethodOrder(MethodSorters.NAME_ASCENDING)
@SuppressWarnings("serial")
public class SmartKeyPlugin {
	
	private static final Logger log = Logger.getLogger(SmartKeyPlugin.class);
	
	private SSHServer sshServer = null;
	
	private static final Map<String, Properties> child_properies = new HashMap<>();
	private static final Properties CONFIG = new Properties();
	
	public static final String KMS_BASE_URL;
	public static final String API_BASE_URL;
	public static final String API_KEY;
	
	private final String secret_key = "mysecret.equinix.com";
	private final String secretKey = "es-engkey";
	private final String secretVal = "equinixdata";
	private final String project_dir = "~/go/src/smartkey-kubernetes-kms";
	private final String config_file = "/etc/smartkey/smartkey-grpc.conf";
	private final String socket_file = "/etc/smartkey/smartkey.socket";
	private final String build_file = "smartkey-kmsplugin_1.0-1_amd64.deb";
	
	static {
		try {
			CONFIG.load(new FileInputStream("./master_config.properties"));
		} catch (Exception e) {
			e.printStackTrace();
			System.exit(1);
		}
		String env = "env_" + getProperty(CONFIG, "ENV") + "_";
		KMS_BASE_URL = getProperty(CONFIG, env + "base_url");
		API_BASE_URL = getProperty(CONFIG, env + "api_url");
		API_KEY = getProperty(CONFIG, env + "api_key");
	}
	
	@Before
	public void setUp() throws Exception {
		String server = "server_" + getProperty(CONFIG, "SERVER") + "_";
		sshServer = new SSHServer(
				getProperty(CONFIG, server + "host"), 
				getProperty(CONFIG, server + "user"), 
				getProperty(CONFIG, server + "password"),
				getProperty(CONFIG, server + "identity_file"));
		sshServer.shellCommand(true, "uname -a");
	}
	
	@After
	public void tearDown() throws Exception {
		sshServer.disconnect();
	}

	/**
	 * Command to test, build or clean project (from inside project). 
	 */
	@Test
	public void _1_TestMakefile() throws Exception {
		sshServer.shellCommand("cd " + project_dir);
		String output = sshServer.shellCommand("ls -1");
		assertTrue("Makefile is missing", output.contains("Makefile")); // Should exist by default
		if (output.contains("smartkey-kms")) {
			sshServer.shellCommand("rm -f smartkey-kms");
			output = sshServer.shellCommand("ls -1");
			assertTrue("smartkey-kms is not removed", !output.contains("smartkey-kms"));
		}
		output = sshServer.shellCommand(true, 3000, "make test | tail -1");
		assertTrue("Make test failed", output.contains("ok  	smartkey-kubernetes-kms"));
		sshServer.shellCommand("make build"); // will create smartkey-kms file
		output = sshServer.shellCommand("ls -1");
		assertTrue("smartkey-kms was not created", output.contains("smartkey-kms"));
		sshServer.shellCommand("make clean");
		output = sshServer.shellCommand("ls -1");
		assertTrue("smartkey-kms must be removed", !output.contains("smartkey-kms"));
	}
	
	/**
	 * Build command will generate smartkey-kms binary
	 */
	@Test
	public void _2_TestCreateInstaller() throws Exception {
		sshServer.shellCommand("sudo rm -rf /etc/smartkey"); //remove old directory
		sshServer.shellCommand("sudo mkdir -p /etc/smartkey");
		sshServer.shellCommand("cd " + project_dir);
		sshServer.shellCommand("rm -f smartkey-kms");
		sshServer.shellCommand("make build"); //create smartkey-kms file
		
		createConfigFile();
		String output = sshServer.shellCommand("sudo ./smartkey-kms --socketFile " + socket_file + " --config " + config_file);
		assertTrue("smartkey-kms is not started", output.contains("KeyManagementServiceServer service started successfully."));
	}
	
	/**
	 * Creating Debian installer from plugin binary
	 */
	@Test
	public void _3_TestCreateDebianInstaller() throws Exception {
		//Should be already installed
		//sshServer.shellCommand("sudo apt-get install dh-make");
		//sshServer.shellCommand("sudo apt-get install devscripts build-essential lintian");
		sshServer.shellCommand("sudo rm -rf " + project_dir + "/debian"); //clear temp directory
		sshServer.shellCommand("rm -f ~/go/src/smartkey-kmsplugin*"); //clean old smartkey build files 
		String output = sshServer.shellCommand("ls " + project_dir + " -1");
		assertTrue("Create Installer script is missing", output.contains("create_installer.sh")); // Should exist by default
		Entry<Integer, List<String>> entry = sshServer.execCommand("cd " + project_dir + "; sudo ./create_installer.sh"); //create smartkey-kmsplugin_1.0-1_amd64.deb (one level above)
		Integer status = entry.getKey();
		List<String> lines = entry.getValue();
		assertTrue("Installation failed", status == 0);
		assertEquals("Finished running lintian.", lastLine(lines));
		lines = (List<String>)sshServer.execCommand("ls -1 ~/go/src").getValue();
		for (String line : lines) {
			if (line.equals(build_file)) { //verify smartkey-kmsplugin_1.0-1_amd64.deb is created (one level above)
				return;
			}
		}
		fail("Debian file is missing");
	}
	
	/**
	 * Remove all installation files to have clear dir like user and install Plugin from Debian installer
	 */
	@Test
	public void _4_TestPluginInstallation() throws Exception {
		//Uninstall old plugin (if already installed)
		sshServer.execCommand("sudo dpkg -P smartkey-kmsplugin");
		
		//Verify all files are removed (should return error code in terminal)
		Integer status = sshServer.execCommand(false, "ls /usr/bin/smartkey-kms").getKey();
		assertTrue("File exists: /usr/bin/smartkey-kms", status == 2);
		status = sshServer.execCommand(false, "ls /lib/systemd/system/smartkey-grpc.service").getKey();
		assertTrue("File exists: /lib/systemd/system/smartkey-grpc.service", status == 2);
		status = sshServer.execCommand(false, "ls /etc/smartkey/smartkey-grpc.conf").getKey();
		assertTrue("File exists: /etc/smartkey/smartkey-grpc.conf", status == 2); // ESS-622
		status = sshServer.execCommand(false, "ls /etc/smartkey/smartkey.yaml").getKey();
		assertTrue("File exists: /etc/smartkey/smartkey.yaml", status == 2);
		
		//Reset kubernetes configuration
		sshServer.execCommand("sudo kubeadm reset -f"); //Should delete /etc/kubernetes/manifests/kube-apiserver.yaml
		
		status = sshServer.execCommand(false, "ls /etc/kubernetes/manifests/kube-apiserver.yaml").getKey();
		assertTrue("File exists: /etc/kubernetes/manifests/kube-apiserver.yaml", status == 2);
		
		sshServer.execCommand("sudo kubeadm init"); //create /etc/kubernetes/manifests/kube-apiserver.yaml
		status = sshServer.execCommand(false, "ls /etc/kubernetes/manifests/kube-apiserver.yaml").getKey();
		assertTrue("Cannot find /etc/kubernetes/manifests/kube-apiserver.yaml", status == 0);
		
		//## Troubleshooting
		//Fix: Make sure that kubernetes config directory has the same permissions as kubernetes config file, ESS-629
		sshServer.execCommand("mkdir -p $HOME/.kube");
		sshServer.execCommand("sudo cp  /etc/kubernetes/admin.conf $HOME/.kube/config"); //Copy new admin certificate into home directory
		sshServer.execCommand("sudo chown $(id -u):$(id -g) $HOME/.kube/config");
		sshServer.execCommand("export KUBECONFIG=$HOME/.kube/config");

		Entry<Integer, List<String>> entry = sshServer.execCommand("sudo dpkg -i go/src/" + build_file);  //install plugin
		status = entry.getKey();
		assertTrue("Debian Installation failed", status == 0);
		//verify that following files got installed
		status = sshServer.execCommand(false, "ls /usr/bin/smartkey-kms").getKey();
		assertTrue("Cannot find /usr/bin/smartkey-kms", status == 0);
		status = sshServer.execCommand(false, "ls /lib/systemd/system/smartkey-grpc.service").getKey();
		assertTrue("Cannot find /lib/systemd/system/smartkey-grpc.service", status == 0);
		status = sshServer.execCommand(false, "ls /etc/smartkey/smartkey-grpc.conf").getKey();
		assertTrue("Cannot find /etc/smartkey/smartkey-grpc.conf", status == 0); // ESS-622
		status = sshServer.execCommand(false, "ls /etc/smartkey/smartkey.yaml").getKey();
		assertTrue("Cannot find /etc/smartkey/smartkey.yaml", status == 0);
		
		status = sshServer.execCommand("kubectl get nodes").getKey();
		log.debug("kubectl get nodes exit code: " + status);
		assertTrue("Cannot read kubernetes nodes", status == 0);
		
		// hexdump convert to binary file (but you can read the text there). Our plugin convert text to SmartKey that nobody can read
		String output = verifyEnctyption();
		assertTrue("Data is not encrypted", output.contains(secretKey) && output.contains(secretVal));
	}
	
	/**
	 * Configure api-server to use "SmartKey Kubernetes KMS Plugin"
	 */
	@Test
	public void _5_TestPluginConfiguration() throws Exception {
		createConfigFile(); //  Update configuration at "/etc/smartkey/smartkey-grpc.conf". Refer above section for more details.
		sshServer.execCommand(false, "sudo service smartkey-grpc start > /dev/null &"); // Execute below command to run plugin gRPC server 
		List<String> lines = sshServer.execCommand("sudo service smartkey-grpc status").getValue(); //Verify if service started correctly using below command
		assertTrue("Plugin service status", lastLine(lines).contains("KeyManagementServiceServer service started successfully."));
		
		sshServer.shellCommand("cd " + project_dir);
		
		Yaml yaml = new Yaml();
		// ## Configure api-server to use "SmartKey Kubernetes KMS Plugin"
		// Open "/etc/kubernetes/manifests/kube-apiserver.yaml" file and add below configuration.
		lines = sshServer.execCommand(false, "sudo cat /etc/kubernetes/manifests/kube-apiserver.yaml").getValue();
		Map<String, Map<String, List<Map<String, Object>>>> map = yaml.load(String.join("\n", lines));
		List<?> containers = (List<?>) map.get("spec").get("containers");
		assertTrue("YMAL spec/containers is empty", containers.size() > 0);
		@SuppressWarnings("unchecked")
		List<Map<String, Object>> volumeMounts = ((Map<String, List<Map<String, Object>>>) containers.get(0)).get("volumeMounts");
		assertTrue("YMAL spec/containers/volumeMounts is empty", volumeMounts.size() > 0);
		
		@SuppressWarnings("unchecked")
		List<String> commands = ((Map<String, List<String>>) containers.get(0)).get("command");
		boolean encryptionExists = commands.contains("--encryption-provider-config=/etc/smartkey/smartkey.yaml");
		if (!encryptionExists) {
			//first line is: "kube-apiserver"
			commands.add(1, "--encryption-provider-config=/etc/smartkey/smartkey.yaml");
		}
		log.info("encryptionExists: " + encryptionExists);
		
		boolean volumeMountExists = false;
		for (Map<String, Object> volumeMount : volumeMounts) {
			if (volumeMount.get("mountPath").equals("/etc/smartkey")) {
				volumeMountExists = true;
				break;
			}
		}
		log.info("volumeMountExists: " + volumeMountExists);
		if (!volumeMountExists) {
			volumeMounts.add(0, new HashMap<String, Object>() {{
				put("mountPath", "/etc/smartkey");
				put("name", "smartkey-kms");
				put("readOnly", true);
			}});
		}

		List<Map<String, Object>> volumes = map.get("spec").get("volumes");
		boolean volumesExists = false;
		for (Map<String, Object> valume : volumes) {
			if (valume.get("name").equals("smartkey-kms")) {
				volumesExists = true;
				break;
			}
		}
		log.info("volumesExists: " + volumesExists);
		if (!volumesExists) {
			volumes.add(0, new HashMap<String, Object>() {{
				put("hostPath", new HashMap<String, Object>() {{
					put("path", "/etc/smartkey");
					put("type", "DirectoryOrCreate");
				}});
				put("name", "smartkey-kms"); // ESS-621
			}});
		}
		
		//Save on the server on any changes in the original file
		if (!encryptionExists || !volumeMountExists || !volumesExists) {
			DumperOptions options = new DumperOptions();
			options.setDefaultFlowStyle(DumperOptions.FlowStyle.BLOCK);
			String output = new Yaml(options).dump(map);
		    Integer status = sshServer.execCommand(false, "sudo tee /etc/kubernetes/manifests/kube-apiserver.yaml <<EOF\n" + output + "EOF").getKey();
		    assertTrue("Cannot update /etc/kubernetes/manifests/kube-apiserver.yaml", status == 0);
		}
		
		log.debug("Waiting restart kubernetes...");
		Thread.sleep(5000); //waiting when kubernetes will aplly new settings
		Integer status = sshServer.execCommand("kubectl get nodes").getKey();
		log.debug("kubectl get nodes exit code: " + status);
		assertTrue("Cannot read kubernetes nodes", status == 0); // ESS-628
		
		String output = verifyEnctyption();
		assertTrue("Data is encrypted", !(output.contains(secretKey) && output.contains(secretVal)) && output.contains("k8s:enc:kms:v1:smartkey-test"));
	}
	
	private void createConfigFile() throws Exception {
		String[] uuid_iv = APIServices.generateAES();
		String body = "{\\\\n" + 
				"  \\\"smartkeyApiKey\\\": \\\"" + API_KEY + "\\\",\\\\n" + 
				"  \\\"encryptionKeyUuid\\\": \\\"" + uuid_iv[0] + "\\\",\\\\n" + 
				"  \\\"iv\\\": \\\"" + uuid_iv[1] + "\\\",\\\\n" + 
				"  \\\"socketFile\\\": \\\"" + socket_file + "\\\",\\\\n" + 
				"  \\\"smartkeyURL\\\": \\\"" + KMS_BASE_URL + "\\\"\\\\n" + 
				"}";
		sshServer.shellCommand("sudo mkdir -p /etc/smartkey"); //create directory
		sshServer.shellCommand("sudo echo -e " + body + " > ~/smartkey_temp.conf");
		sshServer.shellCommand("sudo mv ~/smartkey_temp.conf " + config_file);
	}
	
	private String verifyEnctyption() throws JSchException, IOException {
		sshServer.execCommand("kubectl delete secret " + secret_key);
		List<String> lines = sshServer.execCommand("kubectl create secret generic " + secret_key + " -n default --from-literal=" 
				+ secretKey + "=" + secretVal).getValue();
		assertTrue("Cannot create a secret key", lastLine(lines).equals("secret/" + secret_key + " created"));
		lines = sshServer.execCommand("kubectl get secrets").getValue();
		assertTrue("Cannot find a new secret key", lastLine(lines).contains(secret_key));
		Entry<Integer, List<String>> entry = sshServer.execCommand("sudo ETCDCTL_API=3 etcdctl --cacert=/etc/kubernetes/pki/etcd/ca.crt " + 
										"--cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt " + 
										"--key=/etc/kubernetes/pki/etcd/healthcheck-client.key " +
										"get /registry/secrets/default/" + secret_key + " | hexdump -v -e '\"%_p\"'");
		assertTrue("Cannot run etcdctl", entry.getKey() == 0);
		return String.join("", entry.getValue());
	}
	
	private String lastLine(List<String> lines) {
		return lines.get(lines.size() - 1);
	}
	
	public static String getProperty(Properties config, String key) {
		Properties override_config = new Properties();
		String path = config.getProperty("property_file");
		if (path != null) {
			if (child_properies.get(path) == null) {
				try {
					override_config.load(new FileInputStream(path));
				} catch (Exception e) {
					e.printStackTrace();
				}
				child_properies.put(path, override_config);
			} else {
				override_config = child_properies.get(path);
			}
			String value = getProperty(override_config, key); //Get property from derived file "property_file"
			if (value != null) {
				return value;
			}
		}
		
		String value = System.getProperty(key); //Get property from command line (-D)
		if (value == null) {
			value = config.getProperty(key); //Get property from the property file
		}
		if (value != null && value.startsWith("$")) {
			String param_name = value.substring(1);
			if (config.getProperty(param_name) != null) {
				value = config.getProperty(param_name); //Get property from the local property variable
			} else {
				value = System.getenv().get(param_name);  //Get property from the System environment
			}
		}
		return value;
	}
}