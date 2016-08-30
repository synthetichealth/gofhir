# Install Go

```
sudo apt install golang-go
```

## Create and Configure $GOPATH

```
mkdir ~/go
export GOPATH=$HOME/go
echo "export GOPATH=$HOME/go" >> ~/.bashrc
```

# Install MongoDB

```
sudo apt install mongodb-server
```

# Clone FHIR server code

```
mkdir -p $GOPATH/src/github.com/intervention-engine
cd $GOPATH/src/github.com/intervention-engine
export https_proxy=http://gatekeeper.mitre.org:80
git clone https://github.com/intervention-engine/fhir.git
unset https_proxy
```

# Build FHIR server code and copy binary to /opt

```
cd fhir
go build
sudo mkdir /opt/fhir
cp ./fhir /opt/fhir
```

# Run the server

```
nohup /opt/fhir/fhir &>/dev/null &
```

## Test it

_NOTE: Port 3001 seems to be blocked, so I tested it locally w/ curl_

First search all patients to ensure I get back a bundle with 0 entries:

```
curl http://localhost:3001/Patient
```

Now download a patient example from FHIR spec and post it to our FHIR server:

```
export http_proxy=http://gatekeeper.mitre.org:80
curl -O http://hl7.org/fhir/2016May/patient-example-f001-pieter.json
unset http_proxy
curl -X POST -H "Content-Type: application/json" -d @patient-example-f001-pieter.json http://localhost:3001/Patient
```

Search again to ensure it posted and persisted

```
curl http://localhost:3001/Patient
```

Delete it

```
curl -X DELETE http://localhost:3001/Patient/57a1f0ab1445d41acb85237d
```
