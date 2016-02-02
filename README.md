# AWS Route53 Utility
##Why?
Why this when I can use the AWS cli? Laziness and because it's a self-contained binary. I needed this functionality when automating host zone creation and updating.

## Current Functionality
Export AWS route53 hosted zones and DNS record set. Each hosted zone is saved as a separate JSON file that contains the hosted zone information and complete dns record set.

##Build
-  Clone the repo
-  Install AWS GO SDK:
  ```
  go get -u github.com/aws/aws-sdk-go
  ```
-  Build:
  ``` go build aws_route53_util.go ```

##Usage
Make sure that AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY enviromental vars have been set
### Export all hosted zones
```
./aws_route53_util --region=us-east-1 export-all
```
### Export single hosted zone to file
```
./aws_route53_util --region=us-east-1 export [domain] [output filename]
./aws_route53_util --region=us-east-1 export example.com /tmp/mydata.json
```
### List all hosted zones and records in human readable format to stdout
```
./aws_route53_util --region=eu-west-1 list
```
### List specified host zone in human readable format to stdout
```
./aws_route53_util --region=eu-west-1 list example.com
```
### Import JSON hosted zone file (Not Implemented yet)
```
./aws_route53_util --region=eu-west-1 import [input filename]
./aws_route53_util --region=eu-west-1 import example.com.json
```
#Todo
- add region support currently defaulting to eu-west-1
- list currently returns incorrect results when invalid hostzone name is supplied, should error out
- check if host zone exists for import
- add import  function to import and recreate records in route53
- better error handling
