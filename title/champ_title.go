package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

type Title struct {
	VIN string `json:"vin"`
	Buyer string `json:"buyer"`
	DealerContact string `json:"dealer_contact"`
	Owner string `json:"owner"`
	DateOfPurchase string `json:"date_of_purchase"`
	State string `json:"state"`
	Status string `json:"status"`
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
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
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

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	titles := []Title{
		Title{Buyer: "John Smith", VIN: "9a93ud9dh302jd822hdk228hfjf2", DealerContact: "814add25-f881-4726-b6ca-bd4370ddcc9a", 
		Owner: "4b4ab849-24bb-4c58-986a-2cc4f22d2d7b", DateOfPurchase: "1549634113", Status: "CREATED"},

		Title{Buyer: "Jane Smith", VIN: "ba93ud9d342fd2jd822hdk228hfjf2", DealerContact: "k91add25-f881-4726-b6ca-bd4370ddcc9a", 
		Owner: "9k4ab849-24bb-4c58-986a-2cc4f22d2d7b", DateOfPurchase: "1549634123", Status: "BUYER_SIGNED"},
	}
	i := 0
	for i < len(titles) {
		fmt.Println("i is ", i)
		titleAsBytes, _ := json.Marshal(titles[i])
		APIstub.PutState("TITLE"+strconv.Itoa(i), titleAsBytes)
		fmt.Println("Added", titles[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	fmt.Printf(function)
	if function == "queryTitleByVIN" {
		return s.queryTitleByVIN(APIstub, args)
	} else if function == "createTitle" {
		return s.createTitle(APIstub, args)
	} else if function == "transferTitle" {
		return s.transferTitle(APIstub, args)
	} else if function == "queryAllTitles" {
		return s.queryAllTitles(APIstub)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "queryTitleByVINDate" {
		return s.queryTitleByVINDate(APIstub, args)
	} else if function == "queryTitleByColor" {
		return s.queryTitleByColor(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract Function Name.")
}

func (s *SmartContract) queryTitleByVINDate(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Expected 2 arguments, received " + strconv.Itoa(len(args)))
	}
	vin := args[0]
	date := args[1]

	key := createTransactionKey(vin, date)
	titleAsBytes, err := APIstub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("query by vin and date")
	fmt.Printf("%x", titleAsBytes)
	return shim.Success(titleAsBytes)
}

func (s *SmartContract) queryTitleByVIN(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Excepted 1 argument, received " + strconv.Itoa(len(args)))
	}

	VIN := args[0]
	fmt.Printf(VIN)
	titleQueryString := fmt.Sprintf("{\"selector\":{\"vin\":\"%s\"}}", VIN)
	titleQueryResults, err := getQueryResultForQueryString(APIstub, titleQueryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(titleQueryResults)
}

func (s *SmartContract) queryTitleByColor(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Excepted 1 argument, received " + strconv.Itoa(len(args)))
	}

	color := args[0]
	titleQueryString := fmt.Sprintf("{\"selector\":{\"color\":\"%s\"}}", color)
	titleQueryResults, err := getQueryResultForQueryString(APIstub, titleQueryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(titleQueryResults)
}

func (s *SmartContract) createTitle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 6 {
		return shim.Error("Expected 6 arguments, received " + strconv.Itoa(len(args)))
	}

	vin := args[0]
	// For simplicity, DateOfPurchase is time of title creation.
	dateOfPurchase := time.Now().String()

	title := &Title{VIN: vin, Buyer: args[1], DealerContact: args[2], Owner: args[3], DateOfPurchase: dateOfPurchase, Status: args[4]}
	titleAsBytes, err := json.Marshal(title)
	if err != nil {
		return shim.Error(err.Error())
	}
	titleKey := createTransactionKey(vin, dateOfPurchase)
	err = APIstub.PutState(titleKey, titleAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (s *SmartContract) transferTitle(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//
	// Creates a new title by VIN (args[0]) to a new owner (args[1])
	// Can get signer of transaction. GetCreator call. 
	//

	titleVin := args[0]
	newOwner := args[1]
	// TODO: Perform lookup based on VIN, calculate hash with VIN and DateOfPurchase, GetState with hash.
	titleAsBytes, err := APIstub.GetState(titleVin)  // Call method to look up by VIN.

	if err != nil {
		return shim.Error("Failed to fetch title: " + err.Error())
	} else if titleAsBytes == nil {
		return shim.Error("Title does not exist.")
	}

	newTitle := Title{}
	err = json.Unmarshal(titleAsBytes, &newTitle)
	if err != nil {
		return shim.Error(err.Error())
	}
	newTitle.Owner = newOwner
	titleJSONasBytes, _ := json.Marshal(newTitle)
	err = APIstub.PutState(titleVin, titleJSONasBytes) // Use hash, do not create new asset (diff key).
 
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferTitle (succeded)")
	return shim.Success(nil)
}

func (s *SmartContract) queryAllTitles(APIstub shim.ChaincodeStubInterface) sc.Response  {
	startKey := ""
	endKey := ""

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		fmt.Printf(string(queryResponse.Value))
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
	fmt.Printf("-queryAllTitles:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
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