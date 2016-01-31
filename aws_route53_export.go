package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"io/ioutil"
)

//Custom Struct containing the HostedZone information and Record sets
type MergedZoneData struct {
	ZoneFileInfo  route53.HostedZone
	ZoneRecordSet route53.ListResourceRecordSetsOutput
}

func main() {

	allzones := getallhostedzones()

	//base param struct intialise
	record_params := &route53.ListResourceRecordSetsInput{}

	for _, res := range allzones.HostedZones {

		fmt.Println("Found HostedZone: ", *res.Name)
		zone := &MergedZoneData{}
		zone.ZoneFileInfo = *res

		record_params.HostedZoneId = res.Id
		zone.ZoneRecordSet = *getdnsrecords(record_params)

		//set params for pagination
		record_params.StartRecordName = zone.ZoneRecordSet.NextRecordName
		record_params.StartRecordType = zone.ZoneRecordSet.NextRecordType

		//check results paginated
		is_trunc := *zone.ZoneRecordSet.IsTruncated

		for is_trunc == true {

			//debug
			//fmt.Println("Paramaters: ", record_params)

			results := &MergedZoneData{}
			results.ZoneRecordSet = *getdnsrecords(record_params)

			//append results
			zone.ZoneRecordSet.ResourceRecordSets = append(zone.ZoneRecordSet.ResourceRecordSets, results.ZoneRecordSet.ResourceRecordSets...)

			record_params.StartRecordName = results.ZoneRecordSet.NextRecordName
			record_params.StartRecordType = results.ZoneRecordSet.NextRecordType

			if !*results.ZoneRecordSet.IsTruncated {
				is_trunc = false
			}
			//debug
			//fmt.Println("Still Truncated: ", is_trunc)
		}
		// write JSON to file
		fmt.Println("Number of records found: ", len(zone.ZoneRecordSet.ResourceRecordSets))
		outputJSONfile(*res.Name+"json", *zone)
	}
}

func getallhostedzones() (resp *route53.ListHostedZonesByNameOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})

	params := &route53.ListHostedZonesByNameInput{
		//DNSName:      aws.String(""),
		//HostedZoneId: aws.String(""),
		MaxItems: aws.String("100"),
	}

	resp, err := svc.ListHostedZonesByName(params)
	if err != nil {
		panic(err)
	}
	return resp
}

func getdnsrecords(params *route53.ListResourceRecordSetsInput) (resp *route53.ListResourceRecordSetsOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})

	resp, err := svc.ListResourceRecordSets(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}
	return resp
}

func outputJSONfile(filename string, contents MergedZoneData) {
	output, err := json.MarshalIndent(contents, "", " ")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = ioutil.WriteFile(filename, output, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
