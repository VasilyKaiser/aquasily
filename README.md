# AQUASILY
 A Tool for Domain Flyovers - my version of michenriksen's [aquatone](https://github.com/michenriksen/aquatone)

Aquasily is a tool for visual inspection of websites across a large amount of hosts and is convenient for quickly gaining an overview of HTTP-based attack surface.

* * *

## Installation

1. Install [Google Chrome](https://www.google.com/chrome/) or [Chromium](https://www.chromium.org/getting-involved/download-chromium) or [Edge](https://www.microsoft.com/en-us/edge) browser -- **Note:** If you want to use Edge, you need to pass its path to aquasily using `-browser` switch.
2. Download the latest release of Aquasily for your operating system.
3. Uncompress the zip file and move the `aquasily` binary to your desired location. You probably want to move it to a location in your `$PATH` for easier use.

### Compiling the source code

Comming soon...

* * *

## Usage

Pipe host names to aquasily or URLs if you don't want common web ports scanning.

### Command-line options

| Flag | Description | Default | Example |
| ---- | ----------- | ------- | ------- |
| -version | Print current Aquasily version | `false` | `aquasily -version` |
| -out | Directory to write report to | `.` | `cat hosts.txt | aquasily -out /var/tmp/` |
| -threads | Number of concurrent threads | Number of logical CPUs | `cat hosts.txt | aquasily -threads 20` |
| -ports | Ports to scan on hosts. Supported list aliases: small, medium, large, xlarge | `medium` | `cat hosts.txt | aquasily -ports 80,443,3000,3001` |
| -scan-timeout | Timeout in milliseconds for port scans | `600` | `cat hosts.txt | aquasily -scan-timeout 1500`
| -nmap | Parse input as Nmap/Masscan XML | `false` | `cat scan.xml | aquasily -nmap` |
| browser | Full path to the Chrome/Chromium/Edge executable to use. By default, aquasily will search for Chrome or Chromium | Chrome/Chromium | `cat hosts.txt | aquasily -browser "C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"` |
| -resolution | Screenshot resolution | `1200,900` | `cat hosts.txt | aquasily -resolution 1400,1400`
| -proxy | Proxy to use for HTTP requests | `""` | `cat hosts.txt | aquasily -proxy http://127.0.0.1:8080`
| -http-timeout | Timeout in milliseconds for HTTP requests | 3000 | `cat hosts.txt | aquasily -http-timeout 2000`
| -screenshot-timeout |  | `6000` | `cat hosts.txt | aquasily -screenshot-timeout 2000`
| -silent | Suppress all output except for errors | `false` | `cat hosts.txt | aquasily -silent`
| -debug | Print debugging information | `false` | `cat hosts.txt | aquasily -debug`
| -save-body | Save response bodies to files | `true` | `cat hosts.txt | aquasily -save-body`
| -session | Load Aquasily session file and generate HTML report | `""` | `aquasily -session /var/tmp/aquasily_session.json`
| -template | Path to HTML template to use for report | `""` | `cat hosts.txt | aquasily -template /var/tmp/report_template.html`
