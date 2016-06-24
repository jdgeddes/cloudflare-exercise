# Microservices exercise for CloudFlare

##Requirements


Mongo DB

```
$ go get gopkg.in/mgo.v2   
```
```
$ go get gopkg.in/mgo.v2/bson
```

HttpRouter

```
$ go get github.com/julienschmidt/httprouter
```

##Building

For main server:  
```
$ go build server.go
```

For toy external server to notify of certiticate changes:  
```
$ go build external.go
```

##Running the server

```
Usage of ./server:   
  -certhost="": Hostname to notify via HTTP POST of certification activity change    
  -dbhost="localhost": Hostname of Mongo DB server    
  -dbname="customerdb": Database name to query and store customer information    
  -port=8080: Port for the server to run on    
```

## Interfacing with the server

To add a customer with name `<NAME>` and email `<EMAIL>`

```
curl -XPOST -H 'Content-Type: application/json' -d '{"name": "<NAME>", "email":"<EMAIL>"}' http://localhost:8080/customer
```

To delete a customer with email `<EMAIL>`

```
curl -XDELETE http://localhost:8080/customer/<EMAIL>
```

To create a certificate for customer with email `<EMAIL>`, private key `<KEY>`, certificate body `<BODY>`

```
curl -XPOST -H 'Content-Type: application/json' -d '{"email":"<EMAIL>", "key":"<KEY>", "body":"<BODY>", "active":<true/false>}' http://localhost:8080/certificate
```

To get a list of active certificates from customer with email <EMAIL>

```
curl http://localhost:8080/certificate/<EMAIL>
```

To activate/deactive certificate with id <ID>

```
curl -XPUT -H 'Content-Type: application/json' -d '{"active":<true/false>}' http://localhost:8080/certificate/<ID>   
```

