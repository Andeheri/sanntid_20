# Elevator Project

So you wanna ride the elevator?

![The glorious elevator](elevator.jpg)

## Dependencies

### [Install Go](https://go.dev/doc/install)

### Add the hall request assigner executable:

```bash
cd <this directory>/master/assigner
```
For Linux:
```bash
curl -LJO https://github.com/TTK4145/Project-resources/releases/download/v1.1.1/hall_request_assigner

chmod +x hall_request_assigner
```
For Windows:
```powershell
curl -LJO https://github.com/TTK4145/Project-resources/releases/download/v1.1.1/hall_request_assigner.exe
```

## Usage

```bash
chmod +x runElevator.sh
```

compile go project

```bash
go build -o gloriousElevator main.go
```

make sure to have the elevator server running and then 

```bash
./runElevator.sh
```