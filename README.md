# AQUASILY
 A Screenshot Tool for Domain Flyovers - my version of michenriksen's [aquatone](https://github.com/michenriksen/aquatone)

Aquasily is a tool for visual inspection of websites across a large amount of hosts and is convenient for quickly gaining an overview of HTTP-based attack surface.

* * *
## Installation

1. Install [Google Chrome](https://www.google.com/chrome/) or [Chromium](https://www.chromium.org/getting-involved/download-chromium) or [Edge](https://www.microsoft.com/en-us/edge) browser -- **Note:** If you want to use Edge, you need to pass its path to aquasily using `-browser` switch.
2. Download the latest release of Aquasily for your operating system.
3. Uncompress the zip file and move the `aquasily` binary to your desired location. You probably want to move it to a location in your `$PATH` for easier use.

### Compiling the source code

**Note:** Make sure you have required version of [Go](https://golang.org/dl/) installed (probably the latest).

1. `git clone https://github.com/VasilyKaiser/aquasily.git`
2. `cd aquasily`
3. `go mod tidy`
4. `go build .`

* * *
## Usage

Pipe host names to aquasily or URLs if you don't want common web port scanning.

### Command-line options

| Flag | Description | Default | Example |
| ---- | ----------- | ------- | ------- |
| -version | Print current Aquasily version | `false` | `aquasily -version` |
| -out | Directory to write report to | `.` | `cat hosts.txt \| aquasily -out /var/tmp/` |
| -threads | Number of concurrent threads | Number of logical CPUs | `cat hosts.txt \| aquasily -threads 20` |
| -ports | Ports to scan on hosts. Supported list aliases: small, medium, large, xlarge | `medium` | `cat hosts.txt \| aquasily -ports 80,443,3000,3001` |
| -scan-timeout | Timeout in milliseconds for port scans | `600` | `cat hosts.txt \| aquasily -scan-timeout 1500` |
| -nmap | Parse input as Nmap/Masscan XML | `false` | `cat scan.xml \| aquasily -nmap` |
| -browser | Full path to the Chrome/Chromium/Edge executable to use. By default, aquasily will search for Chrome or Chromium | Chrome/Chromium | `cat hosts.txt \| aquasily -browser "C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"` |
| -resolution | Screenshot resolution | `1200,900` | `cat hosts.txt \| aquasily -resolution 1400,1400` |
| -proxy | Proxy to use for HTTP requests | `""` | `cat hosts.txt \| aquasily -proxy http://127.0.0.1:8080` |
| -http-timeout | Timeout in milliseconds for HTTP requests | `3000` | `cat hosts.txt \| aquasily -http-timeout 2000` |
| -screenshot-timeout |  | `6000` | `cat hosts.txt \| aquasily -screenshot-timeout 2000` |
| -silent | Suppress all output except for errors | `false` | `cat hosts.txt \| aquasily -silent` |
| -debug | Print debugging information | `false` | `cat hosts.txt \| aquasily -debug` |
| -save-body | Save response bodies to files | `true` | `cat hosts.txt \| aquasily -save-body` |
| -session | Load Aquasily session file and generate HTML report | `""` | `aquasily -session /var/tmp/aquasily_session.json` |
| -template | Path to HTML template to use for report | `""` | `cat hosts.txt \| aquasily -template /var/tmp/report_template.html` |

### Usage Examples

Aquasily is designed to play nicely with all kinds of tools. Here's some examples:

**Amass DNS enumeration**

[Amass](https://github.com/OWASP/Amass) uses a bunch of OSINT sources as well as active brute-forcing and clever permutations to quickly identify hundreds, if not thousands, of subdomains on a  domain:

```bash
$ amass -active -brute -o hosts.txt -d yahoo.com
alerts.yahoo.com
ads.yahoo.com
am.yahoo.com
- - - SNIP - - -
prd-vipui-01.infra.corp.gq1.yahoo.com
cp103.mail.ir2.yahoo.com
prd-vipui-01.infra.corp.bf1.yahoo.com
$ cat hosts.txt | aquasily
```

There are plenty of other subdomain enumeration tools out there and Aquasily should work just as well with any other tool:

- [Sublist3r](https://github.com/aboul3la/Sublist3r)
- [Subfinder](https://github.com/subfinder/subfinder)
- [Knock](https://github.com/guelfoweb/knock)
- [Gobuster](https://github.com/OJ/gobuster)

* * *
### Giving data to Aquasily

Aquasily is designed to be as easy to use as possible and to integrate with your existing toolset with no or minimal glue. Aquasily is started by piping output of a command into the tool. It doesn't really care how the piped data looks as URLs, domains, and IP addresses will be extracted with regular expression pattern matching. This means that you can pretty much give it output of any tool you use for host discovery.

IPs, hostnames and domain names in the data will undergo scanning for ports that are typically used for web services and transformed to URLs with correct scheme. If the data contains URLs, they are assumed to be alive and do not undergo port scanning.

**Example**
```bash
cat targets.txt | aquasily
```

* * *
## Output

When Aquasily is done processing the target hosts, it has created a bunch of files and folders inside `aquasilyReport_<date>` in current directory:

P.S. It is also possible to set a permanent default output destination by defining an environment variable `AQUASILY_OUT_PATH`, example:
```bash
export AQUASILY_OUT_PATH="~/aquasilyOutput"
```


- **aquasily_report.html:** 
	- An HTML report to open in a browser that displays all the collected screenshots and response headers clustered by similarity.
- **aquasily_urls.txt:** 
	-  A file containing all responsive URLs. Useful for feeding into other tools.
- **aquasily_session.json:** 
	- A file containing statistics and page data. Useful for automation.
- **headers/:**
	- A folder with files containing raw response headers from processed targets
- **html/:**
	- A folder with files containing the raw response bodies from processed targets. If you are processing a large amount of hosts, and don't need this for further analysis, you can disable this with the `-save-body=false` flag to save some disk space.
- **screenshots/:**
	- A folder with PNG screenshots of the processed targets

The output can easily be zipped and shared with others or archived.

* * *
### Nmap or Masscan

Aquasily can make a report on hosts scanned with the [Nmap](https://nmap.org/) or [Masscan](https://github.com/robertdavidgraham/masscan) portscanner. Simply feed Aquasily the XML output and give it the `-nmap` flag to tell it to parse the input as Nmap/Masscan XML:

```bash
cat scan.xml | aquasily -nmap
```
* * *
## Credits

- Big thanks to [Michael Henriksen](https://twitter.com/michenriksen) for his tool [Aquatone](https://github.com/michenriksen/aquatone), which Aquasily codbase is based on. P.S. buy him [a coffee](https://www.buymeacoffee.com/michenriksen).
- Thanks to the great team behind [ProjectDiscovery](https://github.com/projectdiscovery) for their tool [wappalayzergo](https://github.com/projectdiscovery/wappalyzergo) which is used in Aquasily for fingerprinting.