# go-wsl2-host

> v0.2.0 is out, this drops support for `windows.local`, if this was important let me know and I can add it back in.

A workaround for accessing the WSL2 VM from the Windows host.

This program installs as a service and runs under the local user account. It automatically updates your Windows hosts file with the WSL2 VM's IP address.

The program uses the name of your distro, modified to be a hostname. For example "Ubuntu-18.04" becomes `ubuntu1804.wsl`. If you have more than one running distro, it will be added as well. When the distro stops it is removed from the host file.

I wrote this for my own use but thought it might be useful for others. It's not perfect but gets the job done for me.

To install and run, download a binary from the releases tab. Place it somewhere like your `Documents/` folder.

Open an **elevated/administrator** command prompt:

```
> .\wsl2host.exe install
Windows Username: <username-you-use-to-login-to-windows>
Windows Password: <password-for-this-user>
```

The program will install a service and start it up. Launch `wsl` then from a `cmd` prompt, run `ping ubuntu1804.wsl`. You can check the Windows hosts file to see what was written. The service will automatically update the IP if the WSL2 VM is stopped and started again.

The Windows hosts file is located at: `C:\Windows\System32\drivers\etc\hosts`

**To remove / uninstall the service:**

_NOTE: Upgrading Windows Insider will remove the service, but not cleanly. To reinstall after upgrading, first make sure you've downloaded the latest version of `wsl2host`, then run `remove` before `install`_

Open an **elevated/administrator** command prompt:

```
> .\wsl2host.exe remove
```

**Specifying aliases**

As of v0.3 you can no specify aliases that point to your WSL2 VM IP. Having `some.client.local`, may be useful in your development workflow.

To do this, create the file `~/.wsl2hosts` in your default WSL2 distro. Host names are space separated:
```
some.client.local my-app.local wsl.local
```

