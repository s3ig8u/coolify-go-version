// Start scp command
go func() {
	defer remoteFile.Close()
	fmt.Fprintf(remoteFile, "C0644 %s %s\n", filepath.Base(remotePath), filepath.Base(remotePath))
	io.Copy(remoteFile, file)
	fmt.Fprint(remoteFile, "\x00")
}() 