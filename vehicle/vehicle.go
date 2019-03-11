package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

type Vehicle struct {
	VIN string `json:"vin"`
	ModelYear string `json:"model_year"`
	Manufacturer string `json:"manufacturer"`
	Model string `json:"model"`
	Color string `json:"color"`
	Mileage float64 `json:"mileage"`
	Condition string `json:"condition"`
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func createTransactionKey(keyArgs ...interface{}) (string) {
	var concatKeys string
	for _, arg := range keyArgs {
		concatKeys = concatKeys + fmt.Sprintf("%v", arg)
	}
	key := sha256.New()
	key.Write([]byte(concatKeys))
	return hex.EncodeToString(key.Sum(nil))
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	if function == "createVehicle" {
		return s.createVehicle(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract Function Name.")
}

func (s *SmartContract) createVehicle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
		if len(args) != 7 {
			return shim.Error("Requires 7 arguments, received " + strconv.Itoa(len(args)))
		}
		vin := args[0]
		// Check for existing vehicle and title.
		titleResponse := APIstub.InvokeChaincode("titlecc", [][]byte{[]byte("queryTitleByVIN"), []byte(vin)}, "")
		titlePayload := titleResponse.GetPayload()
		// View errors/debug with `fmt.Printf("%s", titleResponse.GetPayload())``
		titlePayload = titlePayload[1:len(titlePayload) - 1]
		if len(titlePayload) != 0 {
			title := make(map[string]interface{})
			err := json.Unmarshal(titlePayload, &title)
			if err != nil {
				return shim.Error(err.Error())
			}
			fmt.Printf("There is an existing title with this VIN")
		}

		modelYear := args [1]
		manufacturer := args[2]
		model := args[3]
		vehicleKey := createTransactionKey(vin, modelYear, manufacturer, model)

		mileage, err := strconv.ParseFloat(args[4], 64)
		if err != nil {
			return shim.Error(err.Error())
		}
		vehicle := &Vehicle{VIN: vin, ModelYear: modelYear, Manufacturer: manufacturer, Model: model, Mileage: mileage, Condition: args[5], Color: args[6]}
		vehicleAsBytes, err := json.Marshal(vehicle)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = APIstub.PutState(vehicleKey, vehicleAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		// // Index titles on VIN.
		// indexName := "vin~owner"
		// titleBuyerIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{title.VIN, title.Owner})
		// if err != nil {
		// 	return shim.Error(err.Error())
		// }
		// value := []byte{0x00}
		// APIstub.PutState(titleBuyerIndexKey, value)
		return shim.Success(nil)
}

var logger = shim.NewLogger("logger")

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
	logLevel, _ := shim.LogLevel(os.Getenv("FABRIC_LOGGING_SPEC"))
	shim.SetLoggingLevel(logLevel)
}
