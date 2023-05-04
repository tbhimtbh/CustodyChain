package chaincode

import (
        "encoding/json"
        "fmt"

        "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
        contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	CustodianName  string `json:"custodianName"`
	CustodianAgency string `json:"custodianAgency"`
	CaseNumber string `json:"caseNumber"`
	EvidenceInfo  string `json:"evidenceInfo"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
        assets := []Asset{
		{CustodianName: "Zaki", CustodianAgency: "RCED", CaseNumber: "1", EvidenceInfo: "HP01/HP02"},
		{CustodianName: "Aya", CustodianAgency: "RBPF", CaseNumber: "2", EvidenceInfo: "HP01/HP02/HP03"},
		{CustodianName: "Adi", CustodianAgency: "KDN", CaseNumber: "3", EvidenceInfo: "HP01/HP02/SIM01/SIM02"},
		{CustodianName: "Dan", CustodianAgency: "CSB", CaseNumber: "4", EvidenceInfo: "HP01/HP02/SIM01/"},
		{CustodianName: "Azmi", CustodianAgency: "RCED", CaseNumber: "5", EvidenceInfo: "HP01/HP02/HP03/SIM01/SIM02"},
		{CustodianName: "Mirul", CustodianAgency: "CSB", CaseNumber: "6", EvidenceInfo: "HP01/HP02/HP03/SIM01/SIM02/SIM03"},
	}

        for _, asset := range assets {
                assetJSON, err := json.Marshal(asset)
                if err != nil {
                        return err
                }

                err = ctx.GetStub().PutState(asset.CustodianName, assetJSON)
                if err != nil {
                        return fmt.Errorf("failed to put to world state. %v", err)
                }
        }

        return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, custodianName string, custodianAgency string, caseNumber string, evidenceInfo string) error {
        exists, err := s.AssetExists(ctx, custodianName)
        if err != nil {
                return err
        }
        if exists {
                return fmt.Errorf("the asset %s already exists", custodianName)
        }

        asset := Asset{
                CustodianName:        custodianName,
                CustodianAgency:      custodianAgency,
                CaseNumber:           caseNumber,
                EvidenceInfo:         evidenceInfo,
        }
        assetJSON, err := json.Marshal(asset)
        if err != nil {
                return err
        }

        return ctx.GetStub().PutState(custodianName, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given custodianName.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, custodianName string) (*Asset, error) {
        assetJSON, err := ctx.GetStub().GetState(custodianName)
        if err != nil {
                return nil, fmt.Errorf("failed to read from world state: %v", err)
        }
        if assetJSON == nil {
                return nil, fmt.Errorf("the asset %s does not exist", custodianName)
        }

        var asset Asset
        err = json.Unmarshal(assetJSON, &asset)
        if err != nil {
                return nil, err
        }

        return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, custodianName string, custodianAgency string, caseNumber string, evidenceInfo string) error {
        exists, err := s.AssetExists(ctx, custodianName)
        if err != nil {
                return err
        }
        if !exists {
                return fmt.Errorf("the asset %s does not exist", custodianName)
        }

        // overwriting original asset with new asset
        asset := Asset{
                CustodianName:        custodianName,
                CustodianAgency:      custodianAgency,
                CaseNumber:           caseNumber,
                EvidenceInfo:         evidenceInfo,
        }
        assetJSON, err := json.Marshal(asset)
        if err != nil {
                return err
        }

        return ctx.GetStub().PutState(custodianName, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, custodianName string) error {
        exists, err := s.AssetExists(ctx, custodianName)
        if err != nil {
                return err
        }
        if !exists {
                return fmt.Errorf("the asset %s does not exist", custodianName)
        }

        return ctx.GetStub().DelState(custodianName)
}

// AssetExists returns true when asset with given CustodianName exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, custodianName string) (bool, error) {
        assetJSON, err := ctx.GetStub().GetState(custodianName)
        if err != nil {
                return false, fmt.Errorf("failed to read from world state: %v", err)
        }

        return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, caseNumber string, newcustodianName string, newcustodianAgency string) (string, error) {
    asset, err := s.ReadAsset(ctx, caseNumber)
    if err != nil {
        return "", err
    }

    oldcustodianName := asset.CustodianName
    asset.CustodianName = newcustodianName
    asset.CustodianAgency = newcustodianAgency

    assetJSON, err := json.Marshal(asset)
    if err != nil {
        return "", err
    }

    err = ctx.GetStub().PutState(caseNumber, assetJSON)
    if err != nil {
        return "", err
    }

    return oldcustodianName, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
        // range query with empty string for startKey and endKey does an
        // open-ended query of all assets in the chaincode namespace.
        resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
        if err != nil {
                return nil, err
        }
        defer resultsIterator.Close()

        var assets []*Asset
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return nil, err
                }

                var asset Asset
                err = json.Unmarshal(queryResponse.Value, &asset)
                if err != nil {
                        return nil, err
                }
                assets = append(assets, &asset)
        }

        return assets, nil
}
