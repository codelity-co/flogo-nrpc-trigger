package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/golang/protobuf/protoc-gen-go/generator"
)

const (
	serviceName = "\nservice "
)

var registryServerTemplate = template.Must(template.New("").Parse(`// This file registers with nrpc service. This file was auto-generated by mashling at
// {{ .Timestamp }}
package {{.Package}}

import (
	{{if .UnaryMethodInfo}}
	"context"
	"encoding/json"
	{{end}}
	flogoTrigger "github.com/codelity-co/flogo-nrpc-trigger"
)
{{$serviceName := .RegServiceName}}
{{$protoName := .ProtoName}}
{{$option := .Option}}

type serviceImpl{{$protoName}}{{$serviceName}}{{$option}} struct {
	trigger *flogoTrigger.Trigger
	handler *flogoTrigger.Handler
	serviceInfo *flogoTrigger.ServiceInfo
}

var serviceInfo{{$protoName}}{{$serviceName}}{{$option}} = &servInfo.ServiceInfo{
	ProtoName: "{{$protoName}}",
	ServiceName: "{{$serviceName}}",
}

func init() {
	flogoTrigger.ServiceRegistery.RegisterServerService(&serviceImpl{{$protoName}}{{$serviceName}}{{$option}}{
		serviceInfo: serviceInfo{{$protoName}}{{$serviceName}}{{$option}},
	})
}

// RunRegisterServerService registers server method implimentaion with grpc
func (s *serviceImpl{{$protoName}}{{$serviceName}}{{$option}}) RunRegisterServerService(nc *nats.Conn, trigger *flogoTrigger.Trigger, handler *flogoTrigger.Handler) {
	service := &serviceImpl{{$protoName}}{{$serviceName}}{{$option}}{
		trigger: trigger,
		handler: handler,
		serviceInfo: serviceInfo{{$protoName}}{{$serviceName}}{{$option}},
	}
	Register{{$serviceName}}Server(serv, service, handler)
}

func (s *serviceImpl{{$protoName}}{{$serviceName}}{{$option}}) ServiceInfo() *servInfo.ServiceInfo {
	return s.serviceInfo
}

{{- range .UnaryMethodInfo }}

func (s *serviceImpl{{$protoName}}{{$serviceName}}{{$option}}) {{.MethodName}}(ctx context.Context, req *{{.MethodReqName}}) (res *{{.MethodResName}},err error) {
	methodName := "{{.MethodName}}"
	serviceName := "{{$serviceName}}"

	nrpcData := make(map[string]interface{})
	nrpcData["methodName"] = methodName
	nrpcData["serviceName"] = serviceName
	nrpcData["contextData"] = ctx
	nrpcData["reqData"] = req

	s.handler.natsMsgChanngel <- nrpcData

	var reply interface{}
	reply <- s.handler.natsMsgChanngel

	rType := refect.ValueOf(reply).Type()
	if rType  == error {
		return nil, reply
	}
	
	r := &{{.MethodResName}}{}

	replyBytes, err := json.Marshal(reply.(*Reply).Data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(replyBytes, r)
	if err != nil {
		return nil, err
	}
	
	return r, nil
}

{{- end }}

`))

var registryClientTemplate = template.Must(template.New("").Parse(`
`))

var (
	protoPath     string
	protoFileName string
	protoContent  []byte
	appPath       string
	cmdExePath    string
)

// MethodInfoTree holds method information
type MethodInfoTree struct {
	MethodName    string
	MethodReqName string
	MethodResName string
	serviceName   string
}

// ProtoData holds proto file data
type ProtoData struct {
	Timestamp              time.Time
	Package                string
	UnaryMethodInfo        []MethodInfoTree
	ClientStreamMethodInfo []MethodInfoTree
	ServerStreamMethodInfo []MethodInfoTree
	BiDiStreamMethodInfo   []MethodInfoTree
	AllMethodInfo          []MethodInfoTree
	ProtoImpPath           string
	RegServiceName         string
	ProtoName              string
	Option                 string
	Stream                 bool
}

var (
	packageName = flag.String("package", "main", "package name")
)

func main() {
	appPath, _ = os.Getwd()
	flogoJSON := filepath.Join(appPath, "..", "flogo.json")
	_, fileErr := os.Stat(flogoJSON)
	if fileErr != nil {
		// look in parent directory
		flogoJSON = filepath.Join(appPath, "..", "..", "flogo.json")
		_, err := os.Stat(flogoJSON)
		if err != nil {
			log.Println(fmt.Errorf("Cannot find flogo.json file: %s", err.Error()))
		}
	}

	log.Printf("appPath has been set to: %s\n", appPath)

	// Read the flogo.json file to get the proto name and proto content
	inputJSON, err := ioutil.ReadFile(flogoJSON)
	if err != nil {
		log.Println(fmt.Errorf("Error reading json file: %s", err.Error()))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(inputJSON, &data); err != nil {
		log.Println(fmt.Errorf("Error parsing json: %s", err.Error()))
	}

	// parse json to find trigger with protoFile field
	triggers := data["triggers"].([]interface{})
	for _, trigger := range triggers {
		trigger := trigger.(map[string]interface{})
		settings := trigger["settings"].(map[string]interface{})
		if _, ok := settings["protoFile"]; ok {
			if _, okk := settings["protoFile"].(map[string]interface{}); okk {
				// file picker
				protoFile := settings["protoFile"].(map[string]interface{})
				protoFileName = protoFile["filename"].(string)
				protoContent = []byte(protoFile["content"].(string))
				break
			} else {
				// text box
				fmt.Println(os.Getwd())
				protoFileName = settings["protoName"].(string) + ".proto"
				protoContent, err = ioutil.ReadFile(settings["protoFile"].(string))
				if err != nil {
					panic(err)
				}
				break
			}
		}
	}

	// Create a temp proto file with the protoContent
	protoPath = filepath.Join(appPath, protoFileName)
	fmt.Printf("protoPath:[%s] protoFileName:[%s]\n", protoPath, protoFileName)

	err = ioutil.WriteFile(protoPath, protoContent, 0644)
	if err != nil {
		panic(err)
	}

	// Generate support files
	err = GenerateSupportFiles(appPath)
	if err != nil {
		panic(err)
	}

	// cleanup build.go, shim_support.go and <fileName>.proto
	os.Remove(filepath.Join(appPath, "build.go"))
	os.Remove(filepath.Join(appPath, "shim_support.go"))
	os.Remove(protoPath)

	log.Println("Completed build!")

}

// GenerateSupportFiles creates auto genearted code
func GenerateSupportFiles(path string) error {

	log.Println("Getting proto data...")
	pdArr, err := getProtoData()
	if err != nil {
		return err
	}

	// refactoring streaming methods and unary methods
	pdArr = arrangeProtoData(pdArr)

	log.Println("Creating trigger support files...")
	err = generateServiceImplFile(pdArr, "nrpcserver")
	if err != nil {
		return err
	}

	// fmt.Println("Creating client support files...")
	// err = generateServiceImplFile(pdArr, "nrpcclient")
	// if err != nil {
	// 	return err
	// }

	log.Println("Support files created.")
	return nil
}

// getProtoData reads proto and returns proto data present in proto file
func getProtoData() ([]ProtoData, error) {
	var regServiceName string
	var methodInfoList []MethodInfoTree
	var ProtodataArr []ProtoData

	tempString := string(protoContent)
	for i := 0; i < strings.Count(string(protoContent), serviceName); i++ {

		//getting service declaration full string
		tempString = tempString[strings.Index(tempString, serviceName):]

		//getting entire service declaration
		temp := tempString[:strings.Index(tempString, "}")+1]

		regServiceName = strings.TrimSpace(temp[strings.Index(temp, serviceName)+len(serviceName) : strings.Index(temp, "{")])
		regServiceName = generator.CamelCase(regServiceName)
		temp = temp[strings.Index(temp, "rpc")+len("rpc"):]

		//entire rpc methods content
		methodArr := strings.Split(temp, "rpc")

		for _, mthd := range methodArr {
			methodInfo := MethodInfoTree{}
			mthdDtls := strings.Split(mthd, "(")
			methodInfo.MethodName = generator.CamelCase(strings.TrimSpace(mthdDtls[0]))
			methodInfo.MethodReqName = generator.CamelCase(strings.TrimSpace(strings.Split(mthdDtls[1], ")")[0]))
			methodInfo.MethodResName = generator.CamelCase(strings.TrimSpace(strings.Split(mthdDtls[2], ")")[0]))
			methodInfo.serviceName = regServiceName
			methodInfoList = append(methodInfoList, methodInfo)
		}
		protodata := ProtoData{
			Package:        *packageName,
			AllMethodInfo:  methodInfoList,
			Timestamp:      time.Now(),
			ProtoImpPath:   protoPath,
			RegServiceName: regServiceName,
			ProtoName:      strings.Split(protoFileName, ".")[0],
		}

		ProtodataArr = append(ProtodataArr, protodata)
		methodInfoList = nil

		//getting next service content
		tempString = tempString[strings.Index(tempString, serviceName)+len(serviceName):]
	}

	return ProtodataArr, nil
}

// arrangeProtoData refactors different types of methods from all method info list
func arrangeProtoData(pdArr []ProtoData) []ProtoData {

	for index, protoData := range pdArr {
		for _, mthdInfo := range protoData.AllMethodInfo {
			clientStrm := false
			servrStrm := false

			if strings.Contains(mthdInfo.MethodReqName, "stream ") {
				mthdInfo.MethodReqName = strings.Replace(mthdInfo.MethodReqName, "stream ", "", -1)
				clientStrm = true
				protoData.Stream = true
			}
			if strings.Contains(mthdInfo.MethodResName, "stream ") {
				mthdInfo.MethodResName = strings.Replace(mthdInfo.MethodResName, "stream ", "", -1)
				servrStrm = true
				protoData.Stream = true
			}
			if !clientStrm && !servrStrm {
				protoData.UnaryMethodInfo = append(protoData.UnaryMethodInfo, mthdInfo)
			} else if clientStrm && servrStrm {
				protoData.BiDiStreamMethodInfo = append(protoData.BiDiStreamMethodInfo, mthdInfo)
			} else if clientStrm {
				protoData.ClientStreamMethodInfo = append(protoData.ClientStreamMethodInfo, mthdInfo)
			} else if servrStrm {
				protoData.ServerStreamMethodInfo = append(protoData.ServerStreamMethodInfo, mthdInfo)
			}
		}
		pdArr[index] = protoData
	}

	return pdArr
}

// generateServiceImplFile creates implementation files supported for grpc trigger and grpc service
func generateServiceImplFile(pdArr []ProtoData, option string) error {
	dirPath := filepath.Join(appPath)
	_, fileErr := os.Stat(dirPath)
	if fileErr != nil {
		os.MkdirAll(dirPath, os.ModePerm)
	}
	for _, pd := range pdArr {
		connectorFile := filepath.Join(appPath, strings.Split(protoFileName, ".")[0]+"."+pd.RegServiceName+"."+option+".nrpcservice.go")
		f, err := os.Create(connectorFile)
		if err != nil {
			log.Fatal("Error: ", err)
			return err
		}
		defer f.Close()
		pd.Option = option
		if strings.Compare(option, "nrpcserver") == 0 {
			err = registryServerTemplate.Execute(f, pd)
		} else {
			err = registryClientTemplate.Execute(f, pd)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Exec executes a command within the build context.
func Exec(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(cmdExePath) != 0 {
		cmd.Dir = cmdExePath
		cmdExePath = ""
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error executing command: %s \n %s", string(output), err.Error())
	}
	return nil
}
