# journal2awsd
Simple systemd-journal to cloudwatch logs daemon.

This was tested in:
- a local Arch Linux machine running `systemd 243 (243.162-2-arch)`
- Ubuntu EC2 machines running `systemd 237`

The process executes `journalctl` and sends to Cloudwatch logs events every 10 new lines.

# Usage
```bash
$ journal2awsd -h
Usage of journal2awsd:
  -dry-run
    	Set if you want to output messages to console. Useful for testing.
  -group string
    	Specify the log group where you want to send the logs
  -size int
    	Specify the number of events to send to AWS Cloudwatch. (default 10)
  -stream string
    	Specify the log stream where you want to send the logs
  -version
    	Set if you want to see the version and exit.
```

# Run as a service
Two systemd unit files are provided.

- Keep in mind that the binary `journal2awsd` must be available in the original `PATH`.
- Copy the `journal2awsd.service` unit file in the `/etc/systemd/system/` directory.
- Run `# systemctl daemon-reload`
- Run `# systemctl start journal2awsd.service`
- Run `# systemctl status journal2awsd.service`
