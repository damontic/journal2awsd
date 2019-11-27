# journal2awsd
Simple systemd-journal to cloudwatch logs daemon.

This was tested in:
- a local Arch Linux machine running `systemd 243 (243.162-2-arch)`
- Ubuntu EC2 machines running `systemd 237`

The process executes `journalctl` and sends to Cloudwatch logs events every 10 new lines.

# Run as a service
Two systemd unit files are provided.

- Keep in mind that the binary `journal2awsd` must be available in the original `PATH`.
- Copy the `journal2awsd.service` unit file in the `/etc/systemd/system/` directory.
- Run `# systemctl daemon-reload`
- Run `# systemctl start journal2awsd.service`
- Run `# systemctl status journal2awsd.service`
