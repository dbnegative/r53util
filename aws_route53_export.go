package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type MergedZoneData struct {
	ZoneFileInfo  route53.HostedZone
	ZoneRecordSet route53.ListResourceRecordSetsOutput
}

func main() {
	resp := getallhostedzones()

	for _, res := range resp.HostedZones {
		//depointer the string
		fmt.Println("----BEGIN----")
		data := &MergedZoneData{}
		data.ZoneFileInfo = *res
		data.ZoneRecordSet = *getdnsrecords(*res.Id)
		fmt.Println(data)
		fmt.Println("----END----")
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

func getdnsrecords(hostedzoneid string) (resp *route53.ListResourceRecordSetsOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedzoneid), // Required
		MaxItems:     aws.String("100"),
		//StartRecordIdentifier: aws.String("ResourceRecordSetIdentifier"),
		//StartRecordName:       aws.String("DNSName"),
		//StartRecordType:       aws.String("RRType"),
	}
	resp, err := svc.ListResourceRecordSets(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	return resp

}
