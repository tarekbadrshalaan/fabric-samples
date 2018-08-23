package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// PatientsChaincode example simple Patients Chaincode implementation.
type PatientsChaincode struct {
}

// Disease obj.
type Disease struct {
	ID          int64
	Name        string
	Description string
}

// Opration obj.
type Opration struct {
	Description string
	CreatedAt   int64
}

// Patient obj.
type Patient struct {
	ID        int64
	Name      string
	Address   string
	CreatedAt int64
	UpdatedAt int64
	Diseases  []Disease
	History   []Opration
}

// ============================================================
// createPatient - create a new patien, store into chaincode state
// ============================================================
func (t *PatientsChaincode) createPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//	0		1		2
	//	"ID" 	"Name"	"Address"
	//	"1"		"Ali"	"cairo"

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init Patient")
	if len(args[0]) <= 0 {
		return shim.Error("ID: 1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("Name: 2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("Address: 3rd argument must be a non-empty string")
	}

	patientID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	patientIDstr := fmt.Sprintf("Patient-%d", patientID)
	patientName := strings.ToLower(args[1])
	patientAddress := strings.ToLower(args[2])

	// ==== Check if patient already exists ====
	patientAsBytes, err := stub.GetState(patientIDstr)
	if err != nil {
		return shim.Error("Failed to get patient: " + err.Error())
	} else if patientAsBytes != nil {
		msg := fmt.Sprintf("This patient already exists: %v", patientIDstr)
		return shim.Error(msg)
	}

	// ==== Create Patiant object and marshal to JSON ====
	timeNowUNIX := time.Now().Unix()
	patient := &Patient{patientID, patientName, patientAddress, timeNowUNIX, timeNowUNIX, nil, nil}
	patientJSONBytes, err := json.Marshal(patient)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save patient to state ===
	err = stub.PutState(patientIDstr, patientJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the patient to enable Id range queries, e.g. return all Id ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on index:Id.
	//  This will enable very efficient state range queries based on composite keys matching index:Id*
	indexName := "patient~ID"
	IDIndexKey, err := stub.CreateCompositeKey(indexName, []string{"patient", string(patientID)})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the patient.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(IDIndexKey, value)

	// ==== patient saved and indexed. Return success ====
	fmt.Println("- end create patient")
	return shim.Success(nil)
}

// ============================================================
// updatePatientData - change base data for patient
// ============================================================
func (t *PatientsChaincode) updatePatientData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//	0		1		2
	//	"ID" 	"Name"	"Address"
	//	"1"		"Ali"	"cairo"

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// ==== Input sanitation ====
	fmt.Println("- start update Patient")
	if len(args[0]) <= 0 {
		return shim.Error("ID: 1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("Name: 2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("Address: 3rd argument must be a non-empty string")
	}

	patientID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	patientIDstr := fmt.Sprintf("Patient-%d", patientID)
	patientName := strings.ToLower(args[1])
	patientAddress := strings.ToLower(args[2])

	patientAsBytes, err := stub.GetState(patientIDstr)
	if err != nil {
		return shim.Error("Failed to get patient:" + err.Error())
	} else if patientAsBytes == nil {
		return shim.Error("patient does not exist")
	}

	patientToUpdate := Patient{}
	err = json.Unmarshal(patientAsBytes, &patientToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	patientToUpdate.Name = patientName
	patientToUpdate.Address = patientAddress
	patientToUpdate.UpdatedAt = time.Now().Unix()

	patientJSONasBytes, _ := json.Marshal(patientToUpdate)
	err = stub.PutState(patientIDstr, patientJSONasBytes) //rewrite the patient
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end update patient (success)")
	return shim.Success(nil)
}

// ============================================================
// addPatientHistory - add to patient history
// ============================================================
func (t *PatientsChaincode) addPatientHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//	0		1
	//	"ID" 	"history"
	//	"1"		"open heart surgery"

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// ==== Input sanitation ====
	fmt.Println("- start update Patient")
	if len(args[0]) <= 0 {
		return shim.Error("ID: 1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("Histroy: 2nd argument must be a non-empty string")
	}

	patientID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	patientIDstr := fmt.Sprintf("Patient-%d", patientID)
	patientHistory := strings.ToLower(args[1])

	patientAsBytes, err := stub.GetState(patientIDstr)
	if err != nil {
		return shim.Error("Failed to get patient:" + err.Error())
	} else if patientAsBytes == nil {
		return shim.Error("patient does not exist")
	}

	patientToUpdate := Patient{}
	err = json.Unmarshal(patientAsBytes, &patientToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	patientToUpdate.History = append(patientToUpdate.History, Opration{patientHistory, time.Now().Unix()})
	patientToUpdate.UpdatedAt = time.Now().Unix()

	patientJSONasBytes, _ := json.Marshal(patientToUpdate)
	err = stub.PutState(patientIDstr, patientJSONasBytes) //rewrite the patient
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end update patient history (success)")
	return shim.Success(nil)
}

// ============================================================
// getPatientbyID - get Patient full object by Id
// ============================================================
func (t *PatientsChaincode) getPatientbyID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//	0
	//	"ID"
	//	1

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	fmt.Println("- start get Patient")
	if len(args[0]) <= 0 {
		return shim.Error("ID: 1st argument must be a non-empty string")
	}

	patientID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	patientIDstr := fmt.Sprintf("Patient-%d", patientID)

	patientAsBytes, err := stub.GetState(patientIDstr)
	if err != nil {
		return shim.Error("Failed to get patient:" + err.Error())
	} else if patientAsBytes == nil {
		return shim.Error("patient does not exist")
	}

	fmt.Println("- end get patient (success)")
	return shim.Success(patientAsBytes)
}

// ============================================================
// getPatientHistory - get Patient full History by Id
// ============================================================
func (t *PatientsChaincode) getPatientHistorybyID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//	0
	//	"ID"
	//	1

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	fmt.Println("- start get Patient")
	if len(args[0]) <= 0 {
		return shim.Error("ID: 1st argument must be a non-empty string")
	}

	patientID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	patientIDstr := fmt.Sprintf("Patient-%d", patientID)

	pateintHistoryIer, err := stub.GetHistoryForKey(patientIDstr)
	if err != nil {
		return shim.Error("Failed to get patient:" + err.Error())
	} else if pateintHistoryIer == nil {
		return shim.Error("patient does not exist")
	}
	defer pateintHistoryIer.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for pateintHistoryIer.HasNext() {
		response, err := pateintHistoryIer.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

// ============================================================
// createDisease - create a new disease
// ============================================================
func (t *PatientsChaincode) createDisease(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//	0			1				2
	//	"ID" 		"Name" 			"Description"
	//	"1" 		"Diabetes"		"f_Diabetes"

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init disease")
	if len(args[0]) <= 0 {
		return shim.Error("ID: 1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("Name: 2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("Description: 3rd argument must be a non-empty string")
	}

	diseaseID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	diseaseIDstr := fmt.Sprintf("Disease-%d", diseaseID)
	diseaseName := strings.ToLower(args[1])
	diseaseDescription := strings.ToLower(args[2])

	// ==== Check if disease already exists ====
	diseaseAsBytes, err := stub.GetState(diseaseIDstr)
	if err != nil {
		return shim.Error("Failed to get disease: " + err.Error())
	} else if diseaseAsBytes != nil {
		msg := fmt.Sprintf("This Disease already exists: %v", diseaseIDstr)
		return shim.Error(msg)
	}

	// ==== Create disease object and marshal to JSON ====
	disease := &Disease{diseaseID, diseaseName, diseaseDescription}
	diseaseJSONBytes, err := json.Marshal(disease)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save disease to state ===
	err = stub.PutState(diseaseIDstr, diseaseJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the Disease to enable Id range queries, e.g. return all Id ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on index:Id.
	//  This will enable very efficient state range queries based on composite keys matching index:Id*
	indexName := "Disease~ID"
	IDIndexKey, err := stub.CreateCompositeKey(indexName, []string{"Disease", string(diseaseID)})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the patient.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(IDIndexKey, value)

	// ==== disease saved and indexed. Return success ====
	fmt.Println("- end create disease")
	return shim.Success(nil)
}

// ============================================================
// getDiseasebyID - get Disease full object by Id
// ============================================================
func (t *PatientsChaincode) getDiseasebyID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//	0
	//	"ID"
	//	1

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	fmt.Println("- start get Disease")
	if len(args[0]) <= 0 {
		return shim.Error("ID: 1st argument must be a non-empty string")
	}

	diseaseID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	diseaseIDstr := fmt.Sprintf("Disease-%d", diseaseID)

	diseaseAsBytes, err := stub.GetState(diseaseIDstr)
	if err != nil {
		return shim.Error("Failed to get Disease:" + err.Error())
	} else if diseaseAsBytes == nil {
		return shim.Error("Disease does not exist")
	}

	fmt.Println("- end get Disease (success)")
	return shim.Success(diseaseAsBytes)
}

// ============================================================
// assignDiseaseToPatient - assign Disease To Patient
// ============================================================
func (t *PatientsChaincode) assignDiseaseToPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//	0				1
	//	"PatientID" 	"DiseaseID"
	//	"1"				"7"

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// ==== Input sanitation ====
	fmt.Println("- assign Disease To Patient")
	if len(args[0]) <= 0 {
		return shim.Error("PatientID: 1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("DiseaseID: 2nd argument must be a non-empty string")
	}

	// ==== Check if patient exists ====
	patientID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	patientIDstr := fmt.Sprintf("Patient-%d", patientID)

	patientAsBytes, err := stub.GetState(patientIDstr)
	if err != nil {
		return shim.Error("Failed to get patient:" + err.Error())
	} else if patientAsBytes == nil {
		return shim.Error("patient does not exist")
	}

	// ==== Check if disease exists ====
	diseaseID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return shim.Error("ID: 1st argument must be a numeric 64bit of string")
	}
	diseaseIDstr := fmt.Sprintf("Disease-%d", diseaseID)

	diseaseAsBytes, err := stub.GetState(diseaseIDstr)
	if err != nil {
		return shim.Error("Failed to get disease: " + err.Error())
	} else if diseaseAsBytes == nil {
		msg := fmt.Sprintf("Disease does not exist")
		return shim.Error(msg)
	}

	patientToUpdate := Patient{}
	err = json.Unmarshal(patientAsBytes, &patientToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	diseaseToUpdate := Disease{}
	err = json.Unmarshal(diseaseAsBytes, &diseaseToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}

	for _, dis := range patientToUpdate.Diseases {
		if dis.ID == diseaseToUpdate.ID {
			msg := fmt.Sprintf("This Patient:%v-%v already have this Disease:%v-%v", patientToUpdate.ID, patientToUpdate.Name, diseaseToUpdate.ID, diseaseToUpdate.Name)
			return shim.Error(msg)
		}
	}

	patientToUpdate.Diseases = append(patientToUpdate.Diseases, diseaseToUpdate)
	patientToUpdate.UpdatedAt = time.Now().Unix()

	patientJSONasBytes, _ := json.Marshal(patientToUpdate)
	err = stub.PutState(patientIDstr, patientJSONasBytes) //rewrite the patient
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end update patient history (success)")
	return shim.Success(nil)
}

// Init : initialization of chaincode.
func (t *PatientsChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke : Using chaincode (CreatePatients,)
func (t *PatientsChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "createPatient" {
		// Create Patient from patien obj
		return t.createPatient(stub, args)
	} else if function == "updatePatientData" {
		//change base data for patient
		return t.updatePatientData(stub, args)
	} else if function == "addPatientHistory" {
		//add to patient history
		return t.addPatientHistory(stub, args)
	} else if function == "getPatientbyID" {
		//get Patient full object by Id
		return t.getPatientbyID(stub, args)
	} else if function == "getPatientHistorybyID" {
		//get Patient full history by id
		return t.getPatientHistorybyID(stub, args)
	} else if function == "createDisease" {
		//Create New Disease
		return t.createDisease(stub, args)
	} else if function == "getDiseasebyID" {
		//get full Disease object by Id
		return t.getDiseasebyID(stub, args)
	} else if function == "assignDiseaseToPatient" {
		//assign Disease To Patient
		return t.assignDiseaseToPatient(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"createPatient\" \"updatePatientData\" 	\"addPatientHistory\" 	\"getPatientbyID\" 	\"createDisease\" 	\"getDiseasebyID\" 	\"assignDiseaseToPatient\"")
}

func main() {
	err := shim.Start(new(PatientsChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

/*
rm -rf /tmp/heroes-service-* heroes-service
./byfn.sh up
docker exec -it cli bash

/*
	CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer0.org2.example.com:7051 CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt peer chaincode install -n patientssys -v 1.0 -p github.com/chaincode/medical/go

	peer chaincode instantiate -o orderer.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chpatinets -n patientssys -l golang -v 1.0 -c '{"Args":["init"]}' -P 'AND ('\''Org1MSP.peer'\'','\''Org2MSP.peer'\'')'
*/
/*


peer chaincode query -C chpatinets -n patientssys -c '{"Args":["getPatientbyID","1"]}'

//Create
peer chaincode invoke -o orderer.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chpatinets -n patientssys --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["createPatient","1","Ali","cairo"]}'
peer chaincode query -C chpatinets -n patientssys -c '{"Args":["getPatientbyID","1"]}'

//update
peer chaincode invoke -o orderer.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chpatinets -n patientssys --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["updatePatientData","1","Ali Hassan","Alex"]}'
peer chaincode query -C chpatinets -n patientssys -c '{"Args":["getPatientbyID","1"]}'

//addPatientHistory
peer chaincode invoke -o orderer.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chpatinets -n patientssys --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["addPatientHistory","1","open heart surgery"]}'
peer chaincode query -C chpatinets -n patientssys -c '{"Args":["getPatientbyID","1"]}'

//createDisease
peer chaincode invoke -o orderer.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chpatinets -n patientssys --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["createDisease","1","Diabetes","f_Diabetes"]}'
peer chaincode query -C chpatinets -n patientssys -c '{"Args":["getDiseasebyID","1"]}'

//assignDiseaseToPatient
peer chaincode invoke -o orderer.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chpatinets -n patientssys --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"Args":["assignDiseaseToPatient","1","1"]}'
peer chaincode query -C chpatinets -n patientssys -c '{"Args":["getPatientbyID","1"]}'


//getPatientHistory
peer chaincode query -C chpatinets -n patientssys -c '{"Args":["getPatientHistorybyID","1"]}'

*/
