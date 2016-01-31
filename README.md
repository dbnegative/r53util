# AWS route53 utility

### Current Functionality
Export AWS route53 hosted zones and DNS record set. Each hosted zone is saved as a seperate JSON file that contains the hosted zone informatin and complete dns record set.

##Build
- clone the repo
-  Install AWS GO SDK
  ```
  go get -u github.com/aws/aws-sdk-go
  ```
- go build aws_route53_util.go

##Usage
Make sure that AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY enviromental vars have been set
### Export all hosted zones
```
./aws_route53_util export-all
```
### Export single hosted zone
```
./aws_route53_util export [domain] [output filename]
./aws_route53_util export example.com /tmp/mydata.json
```
### Import JSON hosted zone file (Not Implemented yet)
```
./aws_route53_util import [input filename]
./aws_route53_util import example.com.json
```
#Todo 
- add import  function to import and recreate records in route53
- better error handling
- simplified output  
