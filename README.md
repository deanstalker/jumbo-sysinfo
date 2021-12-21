# jumbo-sysinfo

Jumbo sys-info is a small app written to query SMBios, as well as Get-Disk (Windows) and lsblk (*nix/bsd) for ata glance system information.

The script dumps the sys info to JSON.

## Executing

*Windows*

Open Powershell / Windows Terminal / etc as an administrator.

```
bin/main-windows-386.exe
```

Alternatively you can use Powershell to convert the result to a table

```
bin/main-windows-386.exe | ConvertFrom-Json
```

*BSD*

Open a terminal

```
sudo bin/main-freebsd-386.exe
```

*Linux*

Open a terminal

```
sudo bin/main-linux-386
```

Optionally, use `jq` to format the result

```
sudo bin/main-linux-386 | jq
```
