
---

**_Update 12-29-2021:_**  

OlxCLI is not under active development at this time, feel free to submit issues or pull requests, however they may not be promptly addressed.

---

# Overview
OlxCLI is an unofficial command line interface (CLI) for ASPEN's Oneliner application. The CLI provides commands for running common fault simulation tasks through the command line or scripts. The commands interface with ASPEN's Oneliner through the provided application programming interface (API) in order to utilize the robust fault simulation algorithms provided.

OlxCLI is designed with core command line tool principles, including support for command [piping](https://en.wikipedia.org/wiki/Pipeline_(Unix)) and correct utilization of stdin/stdout/stderr where applicable.  

OlxCLI is a demonstration project for the [goolx](https://github.com/readpe/goolx) library. If you require more advanced capabilities than OlxCLI supports, you may use [goolx](https://github.com/readpe/goolx) to create custom Go programs to meet your requirements.

## Supported Architectures
OlxCLI is only intended to run on Windows for i386 architectures. This limitation is due to Oneliner's limited support of other operating systems. Since OlxCLI is written in Go it could be expanded to other operating systems and architectures in the future should things change.


# Installation
## Binary
Download the appropriate version for your platform from [OlxAPI Releases](https://github.com/readpe/olxcli/releases). Once downloaded, the binary can be run from anywhere. Ideally you should install the binary somewhere in your `PATH` environment variable, or you can add the installed directory to the `PATH` variable.

## Build and Install from Source (Advanced Install)
Ensure your environment variables for `GOOS` and `GOARCH` are set correctly for the supported Architectures detailed above, for example: 
```bash
GOOS=windows
GOARCH=386
```

Using `Git Bash` on Windows 10: 
```bash
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/readpe/olxcli.git
cd olxcli
GOOS=windows GOARCH=386 go build -o dist/
```

# Basic Usage
## Test Installation
After installing OlxCLI as outlined above, make sure the directory installed to is contained in your `PATH` variable. You can test OlxCLI has been installed correctly by running the the command with the `--help` flag. 
```bash
olxcli --help
```
You should see an output similar to the following:
```txt
olxcli is an unofficial command line interface for ASPEN's Oneliner.

Usage:
  olxcli [flags]
  olxcli [command]

Available Commands:
  busfault    Run bus fault simulations.
  help        Help about any command
  version     Display the current command line application version

Flags:
  -h, --help   help for olxcli
```
You may further test correct integration with your Oneliner installation by running the `version` command with the `--api` flag to obtain the OlxAPI info. This should confirm a correct initialization of the OlxAPI, otherwise it will return an error. 
```bash
olxapi version --api
```

## Documentation
See the [documentation](https://github.com/readpe/olxcli/wiki) wiki for more details on command usage.

# Disclaimer
This project is not endorsed or affiliated with ASPEN Inc.

# Acknowledgements
Thanks to ASPEN Inc. for providing the Python API which inspired the development of the [goolx](https://github.com/readpe/goolx) library, which was utilized in the development of this CLI.
