package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

var flagRegion string

//ZoneData - Custom Struct containing the HostedZone information and Record sets
type ZoneData struct {
	HostedZone       []*route53.HostedZone
	HostedZoneParams *route53.ListHostedZonesByNameInput
	RecordSets       []*route53.ListResourceRecordSetsOutput
	RecordSetsParams []*route53.ListResourceRecordSetsInput
}

func main() {

	//get cmdline args
	arg := os.Args
	argLength := len(arg)

	//intialise
	zones := &ZoneData{}

Done:
	for idx, resp := range arg {
		switch resp {
		case "import":
			if argLength == 3 {
				loadJSONFile(arg[2], zones)
				zones.restoreHostedZone()

				//zones.outputJSON()
				break Done
			}
		case "export-all":
			zones.getHostedZones()
			zones.getRecordSets()
			for k := range zones.HostedZone {
				outputJSONFile(*zones.HostedZone[k].Name+"json", *zones)
			}
			break Done
		case "export":
			if argLength == 4 {
				zones.HostedZoneParams = &route53.ListHostedZonesByNameInput{}
				fmt.Println("ARG 2: ", arg[2])
				zones.HostedZoneParams.DNSName = aws.String(arg[2])
				zones.HostedZoneParams.MaxItems = aws.String("1")
				zones.getHostedZones()
				zones.getRecordSets()
				outputJSONFile(arg[3], *zones)
				//exportRecord(arg[2], arg[3]
				break Done
			}
			//exportRecord(arg[2], "")
			break Done
		case "list":
			if argLength == 3 {
				zones.HostedZoneParams = &route53.ListHostedZonesByNameInput{}
				zones.HostedZoneParams.DNSName = aws.String(arg[2])
				zones.HostedZoneParams.MaxItems = aws.String("1")
				zones.getHostedZones()
				zones.getRecordSets()
				zones.outputJSON()
				break Done
			} else {
				zones.getHostedZones()
				zones.getRecordSets()
				zones.outputJSON()
				break Done
			}
		case "help":
			printHelp()
			break Done
		} //end of switch

		if idx == argLength-1 {
			fmt.Println("Error: command or variable incorrect or missing")
			printHelp()
		}
	}
} //main

//printHelp - Prints out help menu
func printHelp() {
	fmt.Println("Usage: r53util [COMMAND] [OPTION] ")
	fmt.Println(" - import [FILENAME]              *Import route53 host zone JSON file ")
	fmt.Println(" - export [ZONENAME] [FILENAME]   *Export route53 host zone to a JSON file ")
	fmt.Println(" - list [OPTIONAL HOSTZONE]       *List all host zones or specific zone details if supplied ")
	fmt.Println(" - export-all                     *Export all route53 host zones to JSON file ")
}

//outputJSONFile - Outputs zone data as JSON to file
func outputJSONFile(filename string, zone ZoneData) {
	output, err := json.MarshalIndent(zone, "", " ")
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

//loadJSONFile - Input JSON file to restore
func loadJSONFile(filename string, zone *ZoneData) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = json.Unmarshal(data, zone)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

//getHostedZones - Get host zone by name - this only should only be called once per query
func (zone *ZoneData) getHostedZones() {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String(flagRegion)})
	resp, err := svc.ListHostedZonesByName(zone.HostedZoneParams)
	if err != nil {
		panic(err)
	}
	//Initialise
	zone.HostedZone = make([]*route53.HostedZone, len(resp.HostedZones))
	for k, v := range resp.HostedZones {
		//Populate
		zone.HostedZone[k] = v

	}
}

//getRecordSets - Get all recordsets from associated host zones
func (zone *ZoneData) getRecordSets() {

	svc := route53.New(session.New(), &aws.Config{Region: aws.String(flagRegion)})

	//Initialise
	zone.RecordSetsParams = make([]*route53.ListResourceRecordSetsInput, len(zone.HostedZone))

	for k := range zone.HostedZone {

		//Initialise if this is the first call
		if len(zone.RecordSets) == 0 {

			zone.RecordSets = make([]*route53.ListResourceRecordSetsOutput, len(zone.HostedZone))

			for i := range zone.HostedZone {
				zone.RecordSets[i] = &route53.ListResourceRecordSetsOutput{}
				zone.RecordSetsParams[i] = &route53.ListResourceRecordSetsInput{}
				zone.RecordSetsParams[i].MaxItems = aws.String("2")
				zone.RecordSetsParams[i].HostedZoneId = zone.HostedZone[i].Id
			}
		}

		resp, err := svc.ListResourceRecordSets(zone.RecordSetsParams[k])

		if err != nil {
			fmt.Println("Error: ", err)
			break
		}

		zone.RecordSets[k] = resp

		//deal with pagination
		for *resp.IsTruncated == true {

			zone.RecordSetsParams[k].StartRecordName = aws.String(*resp.NextRecordName)
			zone.RecordSetsParams[k].StartRecordType = aws.String(*resp.NextRecordType)

			resp, err = svc.ListResourceRecordSets(zone.RecordSetsParams[k])

			if err != nil {
				fmt.Println("Error: ", err)
				break
			}

			zone.RecordSets[k].ResourceRecordSets = append(zone.RecordSets[k].ResourceRecordSets, resp.ResourceRecordSets[0])
		}
	}
}

//restoreHostedZone - Restores a Hosted Zone
func (zone *ZoneData) restoreHostedZone() {

	svc := route53.New(session.New(), &aws.Config{Region: aws.String(flagRegion)})

	params := &route53.CreateHostedZoneInput{
		CallerReference: aws.String("hostedzonecreated2"),     // Required
		Name:            aws.String(*zone.HostedZone[0].Name), // Required
		//DelegationSetId: aws.String(""),
		//HostedZoneConfig: zone.HostedZone[0].Config,
		//VPC:              &route53.VPC{
		//VPCId:     aws.String("VPCId"),
		//VPCRegion: aws.String("VPCRegion"),
		//},
	}

	resp, err := svc.CreateHostedZone(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	fmt.Println(resp)
}

//restoreRecordSet - Restores record set to Zone
func (zone *ZoneData) restoreRecordSet() {
	svc := route53.New(session.New())

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action:            aws.String("ChangeAction"), // Required
					ResourceRecordSet: zone.RecordSets[0].ResourceRecordSets[0],
				},
				// More values...
			},
			Comment: aws.String("ResourceDescription"),
		},
		HostedZoneId: aws.String(*zone.HostedZone[0].Id), // Required
	}
	resp, err := svc.ChangeResourceRecordSets(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}

//outputJSON - Output pretty JSON to stdout
func (zone *ZoneData) outputJSON() {
	for k := range zone.HostedZone {
		hostZoneOutput, _ := json.MarshalIndent(zone.HostedZone[k], "", " ")
		recordSetOutput, _ := json.MarshalIndent(zone.RecordSets[k], "", " ")
		fmt.Println(string(hostZoneOutput))
		fmt.Println(string(recordSetOutput))
	}
}
