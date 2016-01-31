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
```
./aws_route53_util
```
#Todo 
- add import  function to import and recreate records in route53
- add help information
- better error handling
- simplified output  
