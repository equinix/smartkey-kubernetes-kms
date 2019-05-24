package com.equinix.smartkey.kubernetes.kms;

import java.io.BufferedReader;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintStream;
import java.util.AbstractMap;
import java.util.ArrayList;
import java.util.List;
import java.util.Map.Entry;
import java.util.Properties;

import org.apache.log4j.Logger;

import com.jcraft.jsch.Channel;
import com.jcraft.jsch.ChannelExec;
import com.jcraft.jsch.JSch;
import com.jcraft.jsch.JSchException;
import com.jcraft.jsch.Session;

public class SSHServer {

	private Session session = null;
	private Channel shellChannel = null;

	private static final Logger log = Logger.getLogger(SSHServer.class);
	private static final Logger sshLog = Logger.getLogger("sshlog");
	
	public SSHServer(String host, String username, String password, String identity_file) throws JSchException {
		JSch jsch = new JSch();
		session = jsch.getSession(username, host, 22);
		if (password == null && identity_file == null) {
			throw new JSchException("The 'password' or 'identity_file' property must be defined.");
		}
		if (password != null) {
			session.setPassword(password);
		} else {
			if (identity_file.startsWith("$")) {
				identity_file = System.getenv().get(identity_file.substring(1));
			}
			jsch.addIdentity(identity_file);
		}
		
		Properties config = new Properties();
		config.put("StrictHostKeyChecking", "no");
		session.setConfig(config);
		session.connect();
		shellChannel = session.openChannel("shell");// only shell
		shellChannel.connect();
	}
	
	public void disconnect() {
		shellChannel.disconnect();
		session.disconnect();
	}
	
	protected String shellCommand(String command) throws InterruptedException, IOException {
		return shellCommand(false, command);
	}
	
	protected String shellCommand(Boolean output, String command) throws InterruptedException, IOException {
		return shellCommand(output, 1000, command);
	}
	
	protected String shellCommand(Boolean output, int timeout, String command) throws InterruptedException, IOException {
		return shellCommand(output, timeout, new String[] {command});
	}
	
	public String shellCommand(Boolean output, int timeout, String ...commands) throws InterruptedException, IOException {
		log.info("\nshellCommand: " + String.join(", ",  commands));
		sshLog.info("\nshellCommand: " + String.join(", ",  commands));
		PrintStream shellStream = new PrintStream(shellChannel.getOutputStream());
		ByteArrayOutputStream out = new ByteArrayOutputStream();
		shellChannel.setOutputStream(out);
		for (String command : commands) {
			shellStream.println(command);
			shellStream.flush();
		}
		Thread.sleep(timeout);
		String lines = out.toString();
		if (output) {
			log.debug(lines);
		}
		out.close();
		return lines;
	}
	
	public Entry<Integer, List<String>> execCommand(String command) throws JSchException, IOException {
		return execCommand(true, command);
	}
	
	public Entry<Integer, List<String>> execCommand(Boolean output, String command) throws JSchException, IOException {
		return execCommand(output, new String[] {command});
	}

	public Entry<Integer, List<String>> execCommand(Boolean output, String ...commands) throws JSchException, IOException {
		String join = String.join("; ",  commands);
		log.info("\nexecCommand: " + join);
		sshLog.info("\nexecCommand: " + join);
		
		List<String> lines = new ArrayList<String>();
		ChannelExec channel = (ChannelExec) session.openChannel("exec");
		ByteArrayOutputStream bos = new ByteArrayOutputStream();
		channel.setErrStream(bos);
		channel.setCommand(join);
		channel.connect();

		BufferedReader br = new BufferedReader(new InputStreamReader(channel.getInputStream()));
		String line = null;
		while ((line = br.readLine()) != null) {
			lines.add(line);
			if (output) {
				log.debug(line);
			}
		}
		br.close();
		
		try {
			Thread.sleep(500); //delay for getting correct status and flush bos
		} catch (InterruptedException e) {
			e.printStackTrace();
		}
		if (lines.size() == 0) {
			if (bos.size() > 0) {
				String error = new String(bos.toByteArray());
				if (output) {
					log.error(error);
				}
				lines.add(error);
			}
		}
		
		channel.disconnect();
		return new AbstractMap.SimpleEntry<>(channel.getExitStatus(), lines);
	}
}
