# Overview
OlxCLI is an unofficial command line interface (CLI) for ASPEN's Oneliner application. The CLI provides commands for running common fault simulation tasks through the command line or scripts. The commands interface with ASPEN's Oneliner through the provided application programming interface (API) in order to utilize the robust fault simulation algorithms provided.

OlxCLI is designed with core command line tool principles, including support for command [piping](https://en.wikipedia.org/wiki/Pipeline_(Unix)) and correct utilization of stdin/stdout/stderr where applicable.  

OlxCLI utilizes the [goolx](https://github.com/readpe/goolx) library to interface with the Oneliner api. If you require more advanced capabilities than OlxCLI supports, you may use [goolx](https://github.com/readpe/goolx) to create custom Go programs to meet your requirements.

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

## Bus Faults Command (busfault)
The `busfault` or `bf` command will run bus fault simulations on buses within the provided case utilizing the parameters provided with the options flags. 

### _Required Flags_
There are two required flags to run the command, you must provide a input file (`*.olr`) using the `--file` or `-f`, and at least one fault connection using `--conn` or `-c` flags.

#### `-f, --file`
Provide the input file in `*.olr` format.

#### `-c, --conn`
The fault connection code to be simulated. Connection codes are **not** case sensitive. You may specify more than one fault connection using the `--conn` or `-c` flags. This will result in more than one fault connection being simulated per bus. For example the following `-c ABC -c AG -c BCG` will run a three line to ground, single line to ground (Phase A), and two line to ground (Phase B-C) faults.

Only one of each type may be simulated, at most four (4) fault connections can be simulated, one for each of the 3LG, 2LG, 1LG, and LL groups. **If you specify more than one connection for a specific group (e.g. `-c AG -c BG -c CG` ) it cannot be guaranteed which fault connection will be simulated**.

Fault connection codes supported:
| Connection Codes | Fault Simulated   |
| ---------------  | ---------------   |
| ABC              | 3LG fault         | 
| BCG, CBG         | 2LG fault (B-C)   | 
| CAG, ACG         | 2LG fault (C-A)   | 
| ABG, BAG         | 2LG fault (A-B)   | 
| AG               | 1LG fault (A-Gnd) | 
| BG               | 1LG fault (B-Gnd) | 
| CG               | 1LG fault (C-Gnd) | 
| BC, CB           | LL fault (B-C)    | 
| CA, AC           | LL fault (C-A)    | 
| AB, BA           | LL fault (A-B)    | 

*3LG = 3 Line-Ground, 2LG = 2 Line-Ground, 1LG = 1 Line-Ground, LL = Line-Line*

### _Filter Flags_
Filter flags provide means for filtering the buses which bus fault simulations are performed on. 

#### `-a, --area`
#### `-z, --zone`
The `--area` and `--zone` zone flags are straight forward and will only run bus faults on buses with matching area and/or zone numbers. 


#### `--vmin`
#### `--vmax`
The `--vmin` and `--vmax` will only run bus faults on buses with nominal bus voltage in kV that falls between these two values. `vmin <= nominalKV <= vmax`

#### `-e, --expression`
A regular expression pattern which will be matched against the bus name. Only buses whos name is matching the regular expression will be simulated. For more details on the supported regular expression patterns see https://pkg.go.dev/regexp/syntax.

### _Config Flags_
#### `-F, --format`
Specify the output format, currently only supports `csv` value. If `csv` is specified output to the terminal or files will be in csv format. Default is tab delimited output for human readability.  

#### `-o, --output`
Specify the output file. If the filename extension ends in ".csv" the output will automatically be csv formatted overriding the `-F`, `--format` flag if set.

#### `-r, --resistance`
Fault resistance in Ohms. 

#### `-x, --reactance`
Fault reactance in Ohms.

#### `-s, --seq`
Output values as sequential components. 

### Examples
Examples are provided for illustrative purposes, refer to the help menu using `--help` to get the most up to date information on command usage.
 
**Example:** 1LG (Phase A) faults on all buses whos name begins with "N"
```bash
olxcli bf -f SAMPLE09.OLR -c AG -e ^N
```
Results in: 
```
Fault Description                                                 Bus Number  Bus Name     Bus kV  Va_mag (kV)  Va_ang  Vb_mag (kV)  Vb_ang  Vc_mag (kV)  Vc_ang  Ia_mag (A)  Ia_ang  Ib_mag (A)  Ib_ang  Ic_mag (A)  Ic_ang
1. Bus Fault on:           6 NEVADA           132. kV 1LG Type=A  6           NEVADA       132.00  0.00         0.0     76.29        -122.5  75.91        117.4   5797.66     -85.9   0.00        -32.4   0.00        -152.7
1. Bus Fault on:          10 NEW HAMPSHR      33.  kV 1LG Type=A  10          NEW HAMPSHR  33.00   0.00         0.0     18.77        -120.3  18.73        116.6   8864.15     -90.6   0.00        -29.7   0.00        -153.9
```

**Example:** 1LG (Phase C) and 2LG (Phase B-C) faults on all buses whos name ends with "E", and whos nominal voltage is 60 kV or higher
```bash
olxcli bf -f SAMPLE09.OLR  -c CG -c BCG -e E$ --vmin 60 
```
Results in: 
```
Fault Description                                                   Bus Number  Bus Name   Bus kV  Va_mag (kV)  Va_ang  Vb_mag (kV)  Vb_ang  Vc_mag (kV)  Vc_ang  Ia_mag (A)  Ia_ang  Ib_mag (A)  Ib_ang  Ic_mag (A)  Ic_ang
1. Bus Fault on:           5 FIELDALE         132. kV 2LG Type=B-C  5           FIELDALE   132.00  74.92        -3.6    0.00         0.0     0.00         0.0     0.00        -90.0   6666.70     153.2   6661.36     33.4
2. Bus Fault on:           5 FIELDALE         132. kV 1LG Type=C    5           FIELDALE   132.00  74.97        -3.5    75.03        -123.7  0.00         0.0     0.00        -90.0   0.00        100.0   6670.71     33.3
1. Bus Fault on:           4 TENNESSEE        132. kV 2LG Type=B-C  4           TENNESSEE  132.00  85.97        1.9     0.00         0.0     0.00         0.0     0.00        0.0     4060.46     167.9   4116.60     34.1
2. Bus Fault on:           4 TENNESSEE        132. kV 1LG Type=C    4           TENNESSEE  132.00  82.91        -5.0    81.78        -111.2  0.00         0.0     0.00        0.0     0.00        0.0     3690.63     40.6
```

**Example:** 1LG (Phase A) faults on all buses under 60 kV, results presented as sequential components
```bash
olxcli bf -f SAMPLE09.OLR -c ag --vmax 60 --seq
```
```
Fault Description                                                 Bus Number  Bus Name     Bus kV  V0_mag (kV)  V0_ang  V1_mag (kV)  V1_ang  V2_mag (kV)  V2_ang  I0_mag (A)  I0_ang  I1_mag (A)  I1_ang  I2_mag (A)  I2_ang
1. Bus Fault on:          10 NEW HAMPSHR      33.  kV 1LG Type=A  10          NEW HAMPSHR  33.00   5.96         178.3   12.50        -1.8    6.54         178.1   2954.72     -90.6   2954.72     -90.6   2954.72     -90.6
1. Bus Fault on:          11 ROANOKE          13.8 kV 1LG Type=A  11          ROANOKE      13.80   3.02         149.6   5.49         -30.5   2.47         149.3   15840.55    -120.4  15840.55    -120.4  15840.55    -120.4
```

**Example:** Output results to csv file
```bash
olxcli bf -f SAMPLE09.OLR  -c CG -c BCG -e E$ --vmin 60 -o results.csv
```






# Disclaimer
This project is not endorsed or affiliated with ASPEN Inc.

# Acknowledgements
Thanks to ASPEN Inc. for providing the Python API which inspired the development of the [goolx](https://github.com/readpe/goolx) library, which was utilized in the development of this CLI.