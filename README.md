# Hotel Rasch Vacancy Board

Native Go-app voor een hotel-informatiebord op een Raspberry Pi die aan een tv hangt.

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

## Klantinstructie

1. zip uitpakken
2. terminal openen in die map
3. `bash start.sh`
4. afsluiten met `Esc`
