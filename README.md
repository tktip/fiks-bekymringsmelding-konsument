# fiks-bekymringsmelding-konsument
Førsteutgave av integrasjon mot FIKS IO for nedlasting av bekymringsmeldinger. Tjenesten er basert på FIKS IO sin [dokumentasjon](https://ks-no.github.io/fiks-plattform/tjenester/fiksio/) og deres [Java POC](https://github.com/ks-no/fiks-bekymringsmelding-konsument-poc). 

I korte trekk funker tjenesten slik:

Tjenesten abonnerer på en AMQP-kø hos FIKS, og mottar meldinger med bekymringmelding-data. Innholdet i meldingene er krypterte zipfiler med JSON og PDF av bekymringsmelding. Zipfilene dekrypteres vha sertifikat, og innholdet lagres i spesifiserte mapper (se dev_cfg/cfg.yml for spesifikasjon). JSON og PDF lagres i egen mappe, og PDF lagres også i en spesifisert backupmappe i tilfelle den slettes før den burde bli slettet.

NB: I utgangspunktet utviklet for å kjøre Linux, men er også lagt funksjonalitet for å kjøre programmet som en tjeneste i Windows. Windows-tjenesten er ikke garantert å funke like bra som Linux-tjenesten. Make-funksjonalitet er ikke garantert å funke i Windows.

For spørsmål, forbedringsforslag e.l. lag issue eller kontakt postmottak.tip@trondheim.kommune.no

####Bygging av programmet:
#####For linux:
Ved å kjøre `make build` fra rot lagres kjørbar fil i `./bin/amd64/fiks-bekymringsmelding-konsument`. 

Denne kan så kjøres med følgende kommando: `CONFIG=file::/my/config/file.yml ./bin/amd64/fiks-bekymringsmelding-konsument`

#####Windows
Fra rot, kjør `go build cmd/fiks-bekymringsmelding-konsument/main.go`. Deretter kan `main.exe` kjøres. Ved kjøring av exe-fil forventes at config-filsti settes som argument. F.eks. `main.exe dev_cfg/cfg.yml`.

#### Vurderte fremtidige forbedringer:
- Reconnect til FIKS-kø ved feil: Per dags dato avsluttes programmet om køoppkoblingen stanser.
- Nedlasting av dokumenter: Per dags dato hentes meldinger kun fra AMQP-meldingen. Dette har vært vanskelig å få gjort tester på og er lagt på is. Meldinger hvis innhold lagres i dokumentlager lastes med andre ord ikke ned.
- Få makefil til å kjøre ordentlig med GOOS=windows.
- Test Make på windows.
 