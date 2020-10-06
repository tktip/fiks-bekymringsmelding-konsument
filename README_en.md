# fiks-bekymringsmelding-konsument
[Norwegian version](https://github.com/tktip/fiks-bekymringsmelding-konsument/blob/master/README.md) 

First version of FIKS IO integration for downloading reports concerning suspected child abuse (TODO: better wording). 
The service is based on the FIKS IO [docs](https://ks-no.github.io/fiks-plattform/tjenester/fiksio/) and their [Java POC](https://github.com/ks-no/fiks-bekymringsmelding-konsument-poc). 

In short, the service works as follows:

The service subscribes to an AMQP queue maintained by FIKS. FIKS adds messages to this data, which the service picks up. The contents of the messages are encrypted zip files containing a JSON and PDF of a report.
Encrypted zip files are decrypted by way of provided certificate data, and their contents are stored in specified folders (See dev_cfg/cfg.yml for configuration).
JSON and PDF folders are stored in separate folders, and PDFs are also stored in a backup folder in case the original is deleted by accident or too early by case workers.

Note: In general, the software is developed on and for Linux, but functionality has been added to run it as a service in Windows. 
The software is not guaranteed to run as smoothly on Windows as on Linux.
Makefile functions are not guaranteed to work on Windows. 
 
For questions, suggestions or other make an issue or contact TIP at postmottak.tip@trondheim.kommune.no

#### Building and running the software:
##### Linux:
Run `make build` in the root folder and run executable found in `./bin/amd64/fiks-bekymringsmelding-konsument`. 

When running executable, remember to add environment variable `CONFIG`\*. 

Example run: `CONFIG=file::/my/config/file.yml ./bin/amd64/fiks-bekymringsmelding-konsument`

\*The software uses [cfger](https://github.com/tktip/cfger), which tries to resolve variable value based on prefixes. `env::` reads from environment, `file::` reads from file system. Without a tag the immediate value is used.

##### Windows
From root, run `go build cmd/fiks-bekymringsmelding-konsument/main.go`. Afterwards, `main.exe` can be run from root. 
When running the executable, a config file path is expected as an argument. E.g. `main.exe C:\config\cfg.yml`.

#### Future improvements to consider:
- Automatic reconnect on queue disconnect/connect timeout. As of now the software exits if unable to connect or if connection dies.
- Download of documents: Currently reports are assumed to always be present inside AMQP messages, and external messages are not supported. Large messages must be downloaded from the FIKS document store, using an AMQP message header. Currently developing this feature is difficult since there is no way to upload large files using the report system.
- Fix makefile so it works properly with GOOS=windows.
- Test Make on windows.
 
