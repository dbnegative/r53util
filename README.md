# AWS Route53 Utility
##Why?
Why this when I can use the AWS cli? Laziness and because it's a self-contained binary. I needed this functionality when automating host zone creation and updating.

## Current Functionality
* Export AWS Route53 hosted zones and DNS record sets. Each hosted zone and associated record sets are combined as a separate JSON file
* List AWS Route53 zone/zones and associated record sets
* Imports JSON files that have been created with r53util, no other format is supported.

##Build
-  Clone the repo
-  Install AWS GO SDK:
  ```
  go get -u github.com/aws/aws-sdk-go
  ```
-  Build:
  ``` go build r53util.go ```

##Usage
Make sure that AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environmental vars have been set
### Export all hosted zones
```
./r53util  export-all
```
### Export single hosted zone to file
```
./r53util export [domain] [output filename]
./r53util export example.com /tmp/mydata.json
```
### List all hosted zones and records in human readable format to stdout
```
./r53util  list
```
### List specified host zone in JSON format to stdout
```
./r53util  list example.com
```
### Import JSON hosted zone file
Files must come from or follow the format outputted by r53util
```
./r53util import [input filename]
./r53util import example.com.json
```
#Known Issues
* Currently not validating or sanitizing inputs
* Does not error out if zone is not matched/found
* VPC ID association is not supported

#Todo
- List currently returns incorrect results when invalid hostzone name is supplied, should error out
- Check if host zone exists for import
- Better error handling
- More output types
