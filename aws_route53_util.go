package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"io/ioutil"
	"os"
)

//Custom Struct containing the HostedZone information and Record sets
type MergedZoneData struct {
	ZoneFileInfo  route53.HostedZone
	ZoneRecordSet route53.ListResourceRecordSetsOutput
}

func main() {

	arg := os.Args
	arg_length := len(arg)

Done:
	for idx, resp := range arg {
		switch resp {
		case "import":
			if arg_length == 3 {
				restorehostedzone(arg[2])
				break Done
			}
		case "export-all":
			exportrecords()
			break Done
		case "export":
			if arg_length == 4 {
				fmt.Println("Exporting Single HostZone")
				exportrecord(arg[2], arg[3])
				break Done
			}
			exportrecord(arg[2], "")
			break Done
		case "help":
			printhelp()
			break Done
		} //end of switch

		if idx == arg_length-1 {
			fmt.Println("Error: command or variable incorrect or missing")
			printhelp()
		}
	}
}

func printhelp() {
	fmt.Println("Usage: aws_route53_util [COMMAND] [OPTION] ")
	fmt.Println(" - import [FILENAME]              *Import route53 hostzone JSON file ")
	fmt.Println(" - export [ZONENAME] [FILENAME]   *Export route53 hostzone to a JSON file ")
	fmt.Println(" - export-all                     *Export all route53 hostzones to JSON file ")
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

func gethostedzone(hostzonename string) (resp *route53.ListHostedZonesByNameOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})

	params := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(hostzonename),
		//HostedZoneId: aws.String(""),
		MaxItems: aws.String("1"),
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

func inputJSONfile(filename string) (resp []byte) {

	reader, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error: Could not open file: ", filename)
		os.Exit(-1)
	}

	resp, err = ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return resp
}

func exportrecords() {
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

func exportrecord(zonename string, filename string) {

	zone := gethostedzone(zonename)

	//base param struct intialise
	recordset_query_params := &route53.ListResourceRecordSetsInput{}

	mzd := &MergedZoneData{}
	for _, hostzonedata := range zone.HostedZones {
		mzd.ZoneFileInfo = *hostzonedata
		recordset_query_params.HostedZoneId = mzd.ZoneFileInfo.Id
		mzd.ZoneRecordSet = *getdnsrecordset(recordset_query_params)

		//set params for pagination
		recordset_query_params.StartRecordName = mzd.ZoneRecordSet.NextRecordName
		recordset_query_params.StartRecordType = mzd.ZoneRecordSet.NextRecordType

		//check results paginated
		is_trunc := *mzd.ZoneRecordSet.IsTruncated

		for is_trunc == true {

			results := &MergedZoneData{}
			results.ZoneRecordSet = *getdnsrecordset(recordset_query_params)

			//append results
			mzd.ZoneRecordSet.ResourceRecordSets = append(mzd.ZoneRecordSet.ResourceRecordSets, results.ZoneRecordSet.ResourceRecordSets...)

			recordset_query_params.StartRecordName = results.ZoneRecordSet.NextRecordName
			recordset_query_params.StartRecordType = results.ZoneRecordSet.NextRecordType

			if !*results.ZoneRecordSet.IsTruncated {
				is_trunc = false
			}
		}
	}
	fmt.Println("Number of records found: ", len(mzd.ZoneRecordSet.ResourceRecordSets))
	if filename == "" {
		outputJSONfile(*mzd.ZoneFileInfo.Name+"json", *mzd)
	} else {
		outputJSONfile(filename, *mzd)
	}
}

func restorehostedzone(filename string) {

	zonedata := &MergedZoneData{}
	rawcontent := inputJSONfile(filename)

	err := json.Unmarshal(rawcontent, zonedata)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("NOT IMPLEMENTED YET - Creating Hosted Zone: ", *zonedata.ZoneFileInfo.Name)

	//Initialise HostZone struct
	newhostedzone := &route53.CreateHostedZoneInput{}

	//Populate required variables - does not fully support zone file i.e vpc id
	newhostedzone.CallerReference = zonedata.ZoneFileInfo.CallerReference
	newhostedzone.Name = zonedata.ZoneFileInfo.Name

	//svc := route53.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})
}
