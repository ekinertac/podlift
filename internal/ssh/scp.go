package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies a file to the remote server using SCP protocol
func (c *Client) CopyFile(localPath, remotePath string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get file info
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Create remote directory if needed
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
		if _, err := c.Execute(mkdirCmd); err != nil {
			return fmt.Errorf("failed to create remote directory: %w", err)
		}
	}

	// Create SCP session
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Open stdin pipe
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin: %w", err)
	}

	// Start SCP command on remote
	go func() {
		defer stdin.Close()
		
		// Send SCP header
		fmt.Fprintf(stdin, "C0644 %d %s\n", fileInfo.Size(), filepath.Base(remotePath))
		
		// Copy file contents
		io.Copy(stdin, localFile)
		
		// Send termination byte
		fmt.Fprint(stdin, "\x00")
	}()

	// Run SCP command
	scpCmd := fmt.Sprintf("scp -t %s", remotePath)
	if err := session.Run(scpCmd); err != nil {
		return fmt.Errorf("SCP transfer failed: %w", err)
	}

	return nil
}

// CopyFileWithProgress copies a file with progress callback
func (c *Client) CopyFileWithProgress(localPath, remotePath string, progressFn func(int64, int64)) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Get file size for progress
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Create remote directory
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
		c.Execute(mkdirCmd)
	}

	// Create session
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Get stdin pipe
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin: %w", err)
	}

	// Transfer with progress tracking
	go func() {
		defer stdin.Close()

		// Send SCP header
		fmt.Fprintf(stdin, "C0644 %d %s\n", fileSize, filepath.Base(remotePath))

		// Copy with progress
		buf := make([]byte, 32*1024) // 32KB buffer
		var transferred int64

		for {
			n, err := localFile.Read(buf)
			if n > 0 {
				stdin.Write(buf[:n])
				transferred += int64(n)
				
				if progressFn != nil {
					progressFn(transferred, fileSize)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
		}

		// Send termination
		fmt.Fprint(stdin, "\x00")
	}()

	// Run SCP
	scpCmd := fmt.Sprintf("scp -t %s", remotePath)
	if err := session.Run(scpCmd); err != nil {
		return fmt.Errorf("SCP transfer failed: %w", err)
	}

	return nil
}

