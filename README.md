# yeetcd

## local setup

### kubernetes dependency

#### CA
If you need to trust additional CAs beyond the defaults, create a file `trusted-cas.pem` with the CAs contained in it

### Intellij

#### import the project

#### import the parent maven module

### Generate source

#### generate the java protobuf files
```bash
./mvnw clean compile -pl protocol -am
```
