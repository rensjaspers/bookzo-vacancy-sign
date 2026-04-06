# Bookzo Vacancy Sign

Native Go-app voor een hotel-informatiebord op een Raspberry Pi die aan een tv hangt.

## Raspberry Pi install

Aanbevolen klantcommando:

```bash
curl -fsSL https://raw.githubusercontent.com/rensjaspers/bookzo-vacancy-sign/main/deploy/install.sh | CONFIG_URL="https://jouw-vps.example.com/bookzo/config.json" bash
```

Dit script:

1. downloadt de laatste GitHub Release zip
2. pakt die uit in `~/bookzo-vacancy-sign-pi-universal`
3. bewaart `config.json` tussen updates
4. start daarna automatisch `start.sh`

Dezelfde opdracht opnieuw uitvoeren is ook de update-flow.

Als je nog geen externe config-url hebt, kan dit ook:

```bash
curl -fsSL https://raw.githubusercontent.com/rensjaspers/bookzo-vacancy-sign/main/deploy/install.sh | bash
```

Dan wordt een voorbeeldconfig geplaatst in `~/bookzo-vacancy-sign-pi-universal/config.json`. Vul daar eerst de echte waardes in en start daarna:

```bash
bash ~/bookzo-vacancy-sign-pi-universal/current/start.sh
```

Je hoeft `config.json` alleen extern te hosten als je de Pi volledig met een one-liner wilt provisionen. Een bestand op je eigen VPS is prima. Zonder `CONFIG_URL` kun je de config ook gewoon lokaal op de Pi invullen.

## Release maken

De installer downloadt altijd de laatste GitHub Release asset `bookzo-vacancy-sign-pi-universal.zip`.

Maak daarom na een wijziging een nieuwe tag, bijvoorbeeld:

```bash
git tag v1.0.0
git push origin v1.0.0
```

De GitHub Actions workflow bouwt dan automatisch een publieke Pi-zip met voorbeeldconfig.

## Development

Lokaal bouwen en draaien op het platform waarop je op dat moment ontwikkelt:

```bash
make build
make run
```

## Productie

Maak eerst `config.pi.json` vanaf `config.pi.example.json`.

Bouw daarna het klantpakket:

```bash
make package-pi-universal
```

Op macOS is voor dit commando Docker nodig.

De app is bedoeld om handmatig gestart te worden. Er is geen `systemd` of andere service-opzet nodig.
