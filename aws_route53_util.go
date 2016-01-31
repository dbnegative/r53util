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
	export()
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

func getdnsrecordset(params *route53.ListResourceRecordSetsInput) (resp *route53.ListResourceRecordSetsOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})

	resp, err := svc.ListResourceRecordSets(params)

	if err != nil {
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

func export() {
	allzones := getallhostedzones()

	//base param struct intialise
	recordset_query_params := &route53.ListResourceRecordSetsInput{}

	for _, res := range allzones.HostedZones {

		fmt.Println("Found HostedZone: ", *res.Name)
		zone := &MergedZoneData{}
		zone.ZoneFileInfo = *res

		recordset_query_params.HostedZoneId = res.Id
		zone.ZoneRecordSet = *getdnsrecordset(recordset_query_params)

		//set params for pagination
		recordset_query_params.StartRecordName = zone.ZoneRecordSet.NextRecordName
		recordset_query_params.StartRecordType = zone.ZoneRecordSet.NextRecordType

		//check results paginated
		is_trunc := *zone.ZoneRecordSet.IsTruncated

		for is_trunc == true {

			results := &MergedZoneData{}
			results.ZoneRecordSet = *getdnsrecordset(recordset_query_params)

			//append results
			zone.ZoneRecordSet.ResourceRecordSets = append(zone.ZoneRecordSet.ResourceRecordSets, results.ZoneRecordSet.ResourceRecordSets...)

			recordset_query_params.StartRecordName = results.ZoneRecordSet.NextRecordName
			recordset_query_params.StartRecordType = results.ZoneRecordSet.NextRecordType

			if !*results.ZoneRecordSet.IsTruncated {
				is_trunc = false
			}
		}

		// write JSON to file
		fmt.Println("Number of records found: ", len(zone.ZoneRecordSet.ResourceRecordSets))
		outputJSONfile(*res.Name+"json", *zone)
	}
}
