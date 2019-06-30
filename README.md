# go-wsl2-host

A workaround for accessing the WSL2 VM from the Windows host.

This program installs as a service and runs under the local user account. It automatically updates your Windows hosts file with the WSL2 VM's IP address.

The program uses the hostname `wsl.local`

I wrote this for my own use but thought it might be useful for others. It's not perfect but gets the job done for me.

To install and run, download a binary from the releases tab. Place it somewhere like your `Documents/` folder.

Open an **elevated/administrator** command prompt:

```
> .\go-wsl2-host.exe install
Windows Username: <username-you-use-to-login-to-windows>
Windows Password: <password-for-this-user>
```

The program will install a service and start it up. Launch `wsl` then from a `cmd` prompt, run `ping wsl.local`. You can check the Windows hosts file to see what was written. The service will automatically update the IP if the WSL2 VM is stopped and started again.

The Windows hosts file is located at: `C:\Windows\System32\drivers\etc\hosts`
