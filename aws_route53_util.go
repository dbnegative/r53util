package awsroute53util

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

//MergedZoneData - Custom Struct containing the HostedZone information and Record sets
type MergedZoneData struct {
	ZoneFileInfo  route53.HostedZone
	ZoneRecordSet route53.ListResourceRecordSetsOutput
}

var flagRegion string

func init() {
	flag.StringVar(&flagRegion, "region", "", "set AWS region")
}

//func main() {
//	flag.Parse()

//	arg := os.Args
//	argLength := len(arg)

//	if flagRegion != "" {
//	Done:
//		for idx, resp := range arg {
//			switch resp {
//			case "import":
//				if argLength == 3 {
//					restoreHostedZone(arg[2])
//					break Done
//				}
//			case "export-all":
//				exportRecords()
//				break Done
//			case "export":
//				if argLength == 4 {
//					fmt.Println("Exporting Single HostZone")
//					exportRecord(arg[2], arg[3])
//					break Done
//				}
//				exportRecord(arg[2], "")
//				break Done
//			case "list":
//				if argLength == 3 {
//					listZone(arg[2])
//					break Done
//				} else {
//					listZone("")
//					break Done
//				}
//			case "help":
//				printHelp()
//				break Done
//			} //end of switch

//			if idx == argLength-1 {
//				fmt.Println("Error: command or variable incorrect or missing")
//				printHelp()
//			}
//		}
//	} else {
//		fmt.Println("Error: Please set a region i.e --region=eu-west-1")
//	}
//}

func printHelp() {
	fmt.Println("Usage: aws_route53_util --region=[AWS REGION] [COMMAND] [OPTION] ")
	fmt.Println(" - import [FILENAME]              *Import route53 host zone JSON file ")
	fmt.Println(" - export [ZONENAME] [FILENAME]   *Export route53 host zone to a JSON file ")
	fmt.Println(" - list [OPTIONAL HOSTZONE]       *List all host zones or specific zone details if supplied ")
	fmt.Println(" - export-all                     *Export all route53 host zones to JSON file ")
}

func getAllHostedZones() (resp *route53.ListHostedZonesByNameOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String(flagRegion)})

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

func getHostedZone(hostzonename string) (resp *route53.ListHostedZonesByNameOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String(flagRegion)})

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

func getDNSRecordSet(params *route53.ListResourceRecordSetsInput) (resp *route53.ListResourceRecordSetsOutput) {
	svc := route53.New(session.New(), &aws.Config{Region: aws.String(flagRegion)})

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

func getPaginatedResults(hostzone *route53.HostedZone) (zoneoutput *MergedZoneData) {
	//base param struct intialise
	recordsetQueryParams := &route53.ListResourceRecordSetsInput{}
	zone := &MergedZoneData{}

	//zone := &MergedZoneData{}
	zone.ZoneFileInfo = *hostzone

	recordsetQueryParams.HostedZoneId = hostzone.Id
	zone.ZoneRecordSet = *getDNSRecordSet(recordsetQueryParams)

	//set params for pagination
	recordsetQueryParams.StartRecordName = zone.ZoneRecordSet.NextRecordName
	recordsetQueryParams.StartRecordType = zone.ZoneRecordSet.NextRecordType

	//check results paginated
	isTruncated := *zone.ZoneRecordSet.IsTruncated

	for isTruncated == true {

		results := &MergedZoneData{}
		results.ZoneRecordSet = *getDNSRecordSet(recordsetQueryParams)

		//append results
		zone.ZoneRecordSet.ResourceRecordSets = append(zone.ZoneRecordSet.ResourceRecordSets, results.ZoneRecordSet.ResourceRecordSets...)

		recordsetQueryParams.StartRecordName = results.ZoneRecordSet.NextRecordName
		recordsetQueryParams.StartRecordType = results.ZoneRecordSet.NextRecordType

		if !*results.ZoneRecordSet.IsTruncated {
			isTruncated = false
		}
	}
	return zone
}

func listZone(zoneName string) {
	zone := &route53.ListHostedZonesByNameOutput{}
	if zoneName != "" {
		zone = getHostedZone(zoneName)
	} else {
		zone = getAllHostedZones()
	}
	for k := range zone.HostedZones {
		mzd := getPaginatedResults(zone.HostedZones[k])
		printhumanreadable(mzd)
	}
}
func exportRecords() {
	allZones := getAllHostedZones()

	for k, v := range allZones.HostedZones {
		mzd := getPaginatedResults(allZones.HostedZones[k])
		// write JSON to file
		fmt.Println("Found Host Zone: ", *v.Name)
		fmt.Println("Number of records found: ", len(mzd.ZoneRecordSet.ResourceRecordSets))
		outputJSONfile(*v.Name+"json", *mzd)
		fmt.Println("Created file: ", *v.Name+"json")
	}
}

func exportRecord(zonename string, filename string) {

	zone := getHostedZone(zonename)

	for k, v := range zone.HostedZones {
		mzd := getPaginatedResults(zone.HostedZones[k])
		// write JSON to file
		fmt.Println("Found Host Zone: ", *v.Name)
		fmt.Println("Number of records found: ", len(mzd.ZoneRecordSet.ResourceRecordSets))

		if filename == "" {
			outputJSONfile(*mzd.ZoneFileInfo.Name+"json", *mzd)
			fmt.Println("Created file: ", *v.Name+"json")
		} else {
			outputJSONfile(filename, *mzd)
			fmt.Println("Created file: ", filename)
		}
	}
}

func restoreHostedZone(filename string) {

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

	printhumanreadable(zonedata)

	//svc := route53.New(session.New(), &aws.Config{Region: aws.String("eu-west-1")})
}

func printhumanreadable(mzd *MergedZoneData) {
	fmt.Println("---------------------------------------------------------------------------------")
	fmt.Println("Host Zone Name: ", *mzd.ZoneFileInfo.Name, " Host Zone ID: ", *mzd.ZoneFileInfo.Id)
	fmt.Println("Caller Reference: ", *mzd.ZoneFileInfo.CallerReference)
	if mzd.ZoneFileInfo.Config.Comment != nil {
		fmt.Println("Comment: ", *mzd.ZoneFileInfo.Config.Comment)
	}
	if mzd.ZoneFileInfo.Config.PrivateZone != nil {
		fmt.Println("Private Zone: ", *mzd.ZoneFileInfo.Config.PrivateZone)
	}
	fmt.Println("---------------------------------------------------------------------------------")

	for k, v := range mzd.ZoneRecordSet.ResourceRecordSets {
		for _, v1 := range v.ResourceRecords {
			fmt.Println(k+1, *v.Name, *v.Type, *v.TTL, *v1.Value)
		}
	}
}
